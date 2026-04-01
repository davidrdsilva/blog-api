package jobs

import "github.com/davidrdsilva/blog-api/internal/domain/models"

// GenerateCommentsJob carries everything the worker needs to generate AI comments for a post.
type GenerateCommentsJob struct {
	PostID  string
	Title   string
	Content *models.EditorJsContent
}
