package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/davidrdsilva/blog-api/internal/application/jobs"
	"github.com/davidrdsilva/blog-api/internal/domain/models"
	"github.com/davidrdsilva/blog-api/internal/domain/repositories"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/ai"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
)

const commentPromptTemplate = `You are simulating a blog comment section. Given the blog post below, generate exactly 15 comments, one for each of the following personality types: angry, conservative, liberal, religious, toxic, offensive, skeptical, enthusiastic, sarcastic, supportive.

Each comment must sound authentic to that personality. Comments should not exceed 150 words. Do not explain your choices. Comments MUST be in pt-BR (Brazilian Portuguese). Comments should be informal and spontaneous.

For each comment also invent a realistic internet username that fits the personality (e.g. "GrumpyDave92" for angry, "ProfessorWilkins" for academic). Usernames should look like real social media handles: no spaces, may include numbers or underscores, 6–20 characters.

Respond ONLY with a valid JSON array. No markdown, no code fences, no explanation. Use exactly this structure:
[
  {"personality": "angry", "username": "...", "content": "..."},
  {"personality": "conservative", "username": "...", "content": "..."},
  {"personality": "liberal", "username": "...", "content": "..."},
  {"personality": "religious", "username": "...", "content": "..."},
  {"personality": "toxic", "username": "...", "content": "..."},
  {"personality": "skeptical", "username": "...", "content": "..."},
  {"personality": "enthusiastic", "username": "...", "content": "..."},
  {"personality": "academic", "username": "...", "content": "..."},
  {"personality": "sarcastic", "username": "...", "content": "..."},
  {"personality": "empathetic", "username": "...", "content": "..."}
]

Blog post title: %s

Blog post content:
%s`

type commentEntry struct {
	Personality string `json:"personality"`
	Username    string `json:"username"`
	Content     string `json:"content"`
}

// AICommentService generates and persists AI-authored comments for a post.
type AICommentService struct {
	ollamaClient ai.AIClient
	commentRepo  repositories.CommentRepository
	logger       *logging.Logger
}

func NewAICommentService(
	client ai.AIClient,
	commentRepo repositories.CommentRepository,
	logger *logging.Logger,
) *AICommentService {
	return &AICommentService{
		ollamaClient: client,
		commentRepo:  commentRepo,
		logger:       logger,
	}
}

// GenerateAndSave builds a prompt from the job, calls Ollama, and persists all
// generated comments in a single transaction. Called by CommentWorker.
func (s *AICommentService) GenerateAndSave(ctx context.Context, job jobs.GenerateCommentsJob) error {
	text := extractPlainText(job.Content)
	prompt := buildCommentPrompt(job.Title, text)

	raw, err := s.ollamaClient.Generate(ctx, prompt)
	if err != nil {
		return fmt.Errorf("ai generation failed: %w", err)
	}

	entries, err := parseCommentEntries(raw)
	if err != nil {
		return fmt.Errorf("failed to parse ai response: %w", err)
	}

	comments := make([]*models.Comment, 0, len(entries))
	for _, e := range entries {
		author := e.Username
		if author == "" {
			// Fall back to personality name if the model omitted the username
			author = e.Personality
		}
		comments = append(comments, &models.Comment{
			PostID:  job.PostID,
			Author:  author,
			Content: e.Content,
		})
	}

	if err := s.commentRepo.CreateBatch(comments); err != nil {
		return fmt.Errorf("failed to save ai comments: %w", err)
	}

	s.logger.Info("AI comments saved", logging.F("postId", job.PostID), logging.F("count", len(comments)))
	return nil
}

// extractPlainText pulls readable text out of Editor.js block content.
func extractPlainText(content *models.EditorJsContent) string {
	if content == nil {
		return ""
	}

	var sb strings.Builder
	for _, block := range content.Blocks {
		switch block.Type {
		case "paragraph", "header":
			if text, ok := block.Data["text"].(string); ok {
				sb.WriteString(text)
				sb.WriteString("\n\n")
			}
		case "list":
			if items, ok := block.Data["items"].([]interface{}); ok {
				for _, item := range items {
					if s, ok := item.(string); ok {
						sb.WriteString("- ")
						sb.WriteString(s)
						sb.WriteString("\n")
					}
				}
				sb.WriteString("\n")
			}
		case "quote":
			if text, ok := block.Data["text"].(string); ok {
				sb.WriteString("> ")
				sb.WriteString(text)
				sb.WriteString("\n\n")
			}
		case "code":
			if code, ok := block.Data["code"].(string); ok {
				sb.WriteString(code)
				sb.WriteString("\n\n")
			}
		}
	}
	return strings.TrimSpace(sb.String())
}

// buildCommentPrompt assembles the full prompt, truncating content to stay within context limits.
func buildCommentPrompt(title, text string) string {
	const maxTextLen = 2000
	if len(text) > maxTextLen {
		text = text[:maxTextLen] + "..."
	}
	return fmt.Sprintf(commentPromptTemplate, title, text)
}

// parseCommentEntries extracts a JSON array from the model's response.
// It first tries a direct unmarshal, then falls back to scanning for array boundaries,
// since some models prepend prose even when instructed not to.
func parseCommentEntries(raw string) ([]commentEntry, error) {
	var entries []commentEntry

	if err := json.Unmarshal([]byte(raw), &entries); err == nil {
		return entries, nil
	}

	start := strings.Index(raw, "[")
	end := strings.LastIndex(raw, "]")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("no JSON array found in response (first 200 chars): %.200s", raw)
	}

	if err := json.Unmarshal([]byte(raw[start:end+1]), &entries); err != nil {
		return nil, fmt.Errorf("failed to unmarshal extracted JSON array: %w", err)
	}

	return entries, nil
}
