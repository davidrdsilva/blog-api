package ai

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/genai"

	"github.com/davidrdsilva/blog-api/config"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
)

type geminiClient struct {
	client  *genai.Client
	model   string
	timeout time.Duration
	logger  *logging.Logger
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
		logger:  logger,
	}, nil
}

func (c *geminiClient) Generate(ctx context.Context, prompt string) (string, error) {
	c.logger.Debug("Gemini: sending generation request", logging.F("model", c.model))

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	result, err := c.client.Models.GenerateContent(ctx, c.model, genai.Text(prompt), nil)
	if err != nil {
		return "", fmt.Errorf("gemini generation failed: %w", err)
	}

	c.logger.Debug("Gemini: generation completed successfully", logging.F("model", c.model))
	return result.Text(), nil
}
