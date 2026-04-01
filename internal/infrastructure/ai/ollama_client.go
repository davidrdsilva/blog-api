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
)

// OllamaClient is the interface used by AICommentService to call the LLM.
type OllamaClient interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

type ollamaClient struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

func NewOllamaClient(cfg *config.Config) OllamaClient {
	return &ollamaClient{
		baseURL: cfg.Ollama.BaseURL,
		model:   cfg.Ollama.Model,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Ollama.TimeoutSeconds) * time.Second,
		},
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

func (c *ollamaClient) Generate(ctx context.Context, prompt string) (string, error) {
	body, err := json.Marshal(ollamaRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal ollama request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to build ollama request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
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

	return ollamaResp.Response, nil
}
