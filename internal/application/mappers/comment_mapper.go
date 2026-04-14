package mappers

import (
	"time"

	"github.com/davidrdsilva/blog-api/internal/application/dtos"
	"github.com/davidrdsilva/blog-api/internal/domain/models"
)

// brt is Brasilia Time (UTC-3), loaded once at package init.
// pgx returns timestamptz values in UTC; we convert to BRT before formatting
// so API responses match the timezone used by the database session.
var brt = func() *time.Location {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.FixedZone("BRT", -3*60*60)
	}
	return loc
}()

func ToCommentResponse(comment *models.Comment) dtos.CommentResponse {
	return dtos.CommentResponse{
		ID:        comment.ID,
		PostID:    comment.PostID,
		Author:    comment.Author,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt.In(brt).Format(time.RFC3339),
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
