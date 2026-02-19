package mappers

import (
	"time"

	"github.com/davidrdsilva/blog-api/internal/application/dtos"
	"github.com/davidrdsilva/blog-api/internal/domain/models"
)

func ToCommentResponse(comment *models.Comment) dtos.CommentResponse {
	return dtos.CommentResponse{
		ID:        comment.ID,
		PostID:    comment.PostID,
		Author:    comment.Author,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt.Format(time.RFC3339),
	}
}

func ToCommentListResponse(comments []*models.Comment) dtos.CommentListResponse {
	responses := make([]dtos.CommentResponse, len(comments))
	for i, comment := range comments {
		responses[i] = ToCommentResponse(comment)
	}

	return dtos.CommentListResponse{
		Data: responses,
	}
}

func CreateCommentRequestToComment(req dtos.CreateCommentRequest) *models.Comment {
	return &models.Comment{
		PostID:    req.PostID,
		Author:    req.Author,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}
}
