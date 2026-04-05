package ai

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"google.golang.org/genai"

	"github.com/davidrdsilva/blog-api/config"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
)

// maxImages limits how many images are sent per request to keep latency and
// token costs predictable.
const maxImages = 5

type geminiClient struct {
	client     *genai.Client
	model      string
	timeout    time.Duration
	httpClient *http.Client // used to fetch images from MinIO before sending inline
	logger     *logging.Logger
}

// NewGeminiClient creates a Gemini API client. Returns an error if the SDK
// cannot be initialised (e.g. invalid API key format at construction time).
func NewGeminiClient(cfg *config.Config, logger *logging.Logger) (AIClient, error) {
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  cfg.Gemini.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	return &geminiClient{
		client:  client,
		model:   cfg.Gemini.Model,
		timeout: time.Duration(cfg.Gemini.TimeoutSeconds) * time.Second,
		// Separate timeout for image fetching so a slow MinIO doesn't eat into
		// the Gemini generation budget.
		httpClient: &http.Client{Timeout: 10 * time.Second},
		logger:     logger,
	}, nil
}

func (c *geminiClient) Generate(ctx context.Context, req GenerateRequest) (string, error) {
	c.logger.Debug("Gemini: sending generation request",
		logging.F("model", c.model),
		logging.F("images", len(req.ImageURLs)),
	)

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	var result *genai.GenerateContentResponse
	var err error

	if len(req.ImageURLs) > 0 {
		result, err = c.generateMultimodal(ctx, req)
	} else {
		result, err = c.client.Models.GenerateContent(ctx, c.model, genai.Text(req.Prompt), nil)
	}

	if err != nil {
		return "", fmt.Errorf("gemini generation failed: %w", err)
	}

	c.logger.Debug("Gemini: generation completed successfully", logging.F("model", c.model))
	return result.Text(), nil
}

// generateMultimodal fetches each image URL, encodes the bytes as inline data,
// and sends a single multimodal content block to Gemini.
func (c *geminiClient) generateMultimodal(ctx context.Context, req GenerateRequest) (*genai.GenerateContentResponse, error) {
	parts := []*genai.Part{{Text: req.Prompt}}

	urls := req.ImageURLs
	if len(urls) > maxImages {
		c.logger.Debug("Gemini: capping images", logging.F("total", len(urls)), logging.F("cap", maxImages))
		urls = urls[:maxImages]
	}

	for _, url := range urls {
		blob, mimeType, err := c.fetchImage(ctx, url)
		if err != nil {
			// A single failed image is not fatal — skip and continue.
			c.logger.Warn("Gemini: skipping image, fetch failed",
				logging.F("url", url),
				logging.F("error", err.Error()),
			)
			continue
		}
		parts = append(parts, &genai.Part{
			InlineData: &genai.Blob{Data: blob, MIMEType: mimeType},
		})
	}

	contents := []*genai.Content{{Parts: parts, Role: "user"}}
	return c.client.Models.GenerateContent(ctx, c.model, contents, nil)
}

// fetchImage downloads the image at url and returns its raw bytes and MIME type.
// The MIME type is read from the Content-Type response header; it falls back to
// "image/jpeg" when the header is absent or non-specific.
func (c *geminiClient) fetchImage(ctx context.Context, url string) ([]byte, string, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to build image request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, "", fmt.Errorf("image fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("image server returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read image body: %w", err)
	}

	mimeType := resp.Header.Get("Content-Type")
	// Strip parameters like "; charset=utf-8" and fall back for generic values.
	if idx := strings.Index(mimeType, ";"); idx != -1 {
		mimeType = strings.TrimSpace(mimeType[:idx])
	}
	if mimeType == "" || mimeType == "application/octet-stream" {
		mimeType = "image/jpeg"
	}

	return data, mimeType, nil
}
