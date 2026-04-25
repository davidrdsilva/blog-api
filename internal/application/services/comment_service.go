package services

import (
	"fmt"

	"github.com/davidrdsilva/blog-api/config"
	"github.com/davidrdsilva/blog-api/internal/application/dtos"
	"github.com/davidrdsilva/blog-api/internal/application/mappers"
	"github.com/davidrdsilva/blog-api/internal/domain/models"
	"github.com/davidrdsilva/blog-api/internal/domain/repositories"
)

// Matched as a substring by the comment handler to map to WHITENEST_COMMENTS_DISABLED.
const errWhitenestCommentsDisabled = "comments are disabled for whitenest chapters"

type CommentService struct {
	repo     repositories.CommentRepository
	postRepo repositories.PostRepository
	cfg      *config.Config
}

func NewCommentService(
	repo repositories.CommentRepository,
	postRepo repositories.PostRepository,
	cfg *config.Config,
) *CommentService {
	return &CommentService{
		repo:     repo,
		postRepo: postRepo,
		cfg:      cfg,
	}
}

func (s *CommentService) CreateComment(req dtos.CreateCommentRequest) (*dtos.CommentResponse, error) {
	if s.postRepo != nil && req.PostID != "" {
		post, err := s.postRepo.FindByID(req.PostID)
		if err != nil {
			return nil, fmt.Errorf("failed to verify post: %w", err)
		}
		if post != nil && post.WhitenestChapterNumber != nil {
			return nil, fmt.Errorf("%s: post %s is chapter %d",
				errWhitenestCommentsDisabled, post.ID, *post.WhitenestChapterNumber)
		}
	}

	comment := mappers.CreateCommentRequestToComment(req)
	if err := s.repo.Create(comment); err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}
	response := mappers.ToCommentResponse(comment)
	return &response, nil
}

func (s *CommentService) GetComments(filters models.CommentFilters) (*dtos.CommentListResponse, error) {
	comments, err := s.repo.FindAll(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}
	response := mappers.ToCommentListResponse(comments)
	return &response, nil
}

func (s *CommentService) DeleteComment(id string) error {
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}
	return nil
}
