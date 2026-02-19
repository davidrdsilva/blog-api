package services

import (
	"fmt"

	"github.com/davidrdsilva/blog-api/config"
	"github.com/davidrdsilva/blog-api/internal/application/dtos"
	"github.com/davidrdsilva/blog-api/internal/application/mappers"
	"github.com/davidrdsilva/blog-api/internal/domain/models"
	"github.com/davidrdsilva/blog-api/internal/domain/repositories"
)

type CommentService struct {
	repo repositories.CommentRepository
	cfg  *config.Config
}

func NewCommentService(repo repositories.CommentRepository, cfg *config.Config) *CommentService {
	return &CommentService{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *CommentService) CreateComment(req dtos.CreateCommentRequest) (*dtos.CommentResponse, error) {
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
