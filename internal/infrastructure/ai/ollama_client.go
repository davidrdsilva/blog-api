package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/davidrdsilva/blog-api/config"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
)

// GenerateRequest carries the inputs for a generation call.
// ImageURLs is optional and ignored by text-only backends.
type GenerateRequest struct {
	Prompt    string
	ImageURLs []string
}

// AIClient is the interface used by AICommentService to call any LLM backend.
type AIClient interface {
	Generate(ctx context.Context, req GenerateRequest) (string, error)
}

type ollamaClient struct {
	baseURL    string
	model      string
	httpClient *http.Client
	logger     *logging.Logger
}

func NewOllamaClient(cfg *config.Config, logger *logging.Logger) AIClient {
	return &ollamaClient{
		baseURL: cfg.Ollama.BaseURL,
		model:   cfg.Ollama.Model,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Ollama.TimeoutSeconds) * time.Second,
		},
		logger: logger,
	}
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

func (c *ollamaClient) Generate(ctx context.Context, req GenerateRequest) (string, error) {
	c.logger.Debug("Ollama: sending generation request",
		logging.F("model", c.model),
		logging.F("url", c.baseURL),
	)

	body, err := json.Marshal(ollamaRequest{
		Model:  c.model,
		Prompt: req.Prompt,
		Stream: false,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal ollama request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to build ollama request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read ollama response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(rawBody))
	}

	var ollamaResp ollamaResponse
	if err := json.Unmarshal(rawBody, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode ollama response: %w", err)
	}

	if ollamaResp.Error != "" {
		return "", fmt.Errorf("ollama error: %s", ollamaResp.Error)
	}

	c.logger.Debug("Ollama: generation completed successfully", logging.F("model", c.model))
	return ollamaResp.Response, nil
}
