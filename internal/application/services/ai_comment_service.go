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

type commentEntry struct {
	Personality string `json:"personality"`
	Username    string `json:"username"`
	Content     string `json:"content"`
}

// AICommentService generates and persists AI-authored comments for a post.
type AICommentService struct {
	ollamaClient ai.AIClient
	commentRepo  repositories.CommentRepository
	postRepo     repositories.PostRepository
	logger       *logging.Logger
}

func NewAICommentService(
	client ai.AIClient,
	commentRepo repositories.CommentRepository,
	postRepo repositories.PostRepository,
	logger *logging.Logger,
) *AICommentService {
	return &AICommentService{
		ollamaClient: client,
		commentRepo:  commentRepo,
		postRepo:     postRepo,
		logger:       logger,
	}
}

// GenerateAndSave builds a prompt from the job, calls the AI client, and
// persists all generated comments in a single transaction. Called by CommentWorker.
//
// Re-fetches the post to filter out Whitenest chapters so backfills, retries,
// or future enqueue paths can't slip past the dispatcher's skip.
func (s *AICommentService) GenerateAndSave(ctx context.Context, job jobs.GenerateCommentsJob) error {
	if s.postRepo != nil {
		post, err := s.postRepo.FindByID(job.PostID)
		if err != nil {
			return fmt.Errorf("failed to load post for AI comment job: %w", err)
		}
		if post == nil {
			s.logger.Warn("AI comment job dropped: post no longer exists",
				logging.F("postId", job.PostID),
			)
			return nil
		}
		if post.WhitenestChapterNumber != nil {
			s.logger.Info("AI comment job skipped: Whitenest chapter",
				logging.F("postId", job.PostID),
				logging.F("chapter", *post.WhitenestChapterNumber),
			)
			return nil
		}
	}

	text := extractPlainText(job.Content)
	imageURLs := extractImageURLs(job.Content)

	raw, err := s.ollamaClient.Generate(ctx, ai.GenerateRequest{
		Prompt:    buildCommentPrompt(job.Title, text),
		ImageURLs: imageURLs,
	})
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

// extractImageURLs collects the file URL from every Editor.js image block.
func extractImageURLs(content *models.EditorJsContent) []string {
	if content == nil {
		return nil
	}

	var urls []string
	for _, block := range content.Blocks {
		if block.Type != "image" {
			continue
		}
		fileData, ok := block.Data["file"].(map[string]interface{})
		if !ok {
			continue
		}
		url, ok := fileData["url"].(string)
		if !ok || url == "" {
			continue
		}
		urls = append(urls, url)
	}
	return urls
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
