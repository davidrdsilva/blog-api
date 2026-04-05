package ai

import (
	"context"

	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
)

type fallbackAIClient struct {
	primary  AIClient
	fallback AIClient
	logger   *logging.Logger
}

// NewFallbackClient returns an AIClient that tries primary first and, on any
// error, transparently retries with fallback. Both clients must be non-nil.
func NewFallbackClient(primary, fallback AIClient, logger *logging.Logger) AIClient {
	return &fallbackAIClient{
		primary:  primary,
		fallback: fallback,
		logger:   logger,
	}
}

func (c *fallbackAIClient) Generate(ctx context.Context, req GenerateRequest) (string, error) {
	c.logger.Debug("FallbackAI: trying primary client (Gemini)")

	result, err := c.primary.Generate(ctx, req)
	if err != nil {
		c.logger.Warn("FallbackAI: primary client failed, retrying with Ollama",
			logging.F("error", err.Error()),
		)
		return c.fallback.Generate(ctx, req)
	}

	c.logger.Debug("FallbackAI: primary client succeeded, no fallback needed")
	return result, nil
}
