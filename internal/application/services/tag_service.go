package services

import (
	"fmt"

	"github.com/davidrdsilva/blog-api/internal/application/dtos"
	"github.com/davidrdsilva/blog-api/internal/application/mappers"
	"github.com/davidrdsilva/blog-api/internal/domain/models"
	"github.com/davidrdsilva/blog-api/internal/domain/repositories"
)

// TagService handles business logic for tags
type TagService struct {
	repo repositories.TagRepository
}

// NewTagService creates a new tag service
func NewTagService(repo repositories.TagRepository) *TagService {
	return &TagService{repo: repo}
}

func (s *TagService) ListTags(filters models.TagFilters) (*dtos.TagListResponse, error) {
	tags, err := s.repo.FindAll(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	resp := mappers.ToTagListResponse(tags)
	return &resp, nil
}
