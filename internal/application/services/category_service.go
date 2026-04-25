package services

import (
	"fmt"

	"github.com/davidrdsilva/blog-api/internal/application/dtos"
	"github.com/davidrdsilva/blog-api/internal/application/mappers"
	"github.com/davidrdsilva/blog-api/internal/domain/models"
	"github.com/davidrdsilva/blog-api/internal/domain/repositories"
)

// CategoryService handles business logic for categories
type CategoryService struct {
	repo repositories.CategoryRepository
}

// NewCategoryService creates a new category service
func NewCategoryService(repo repositories.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) ListCategories(filters models.CategoryFilters) (*dtos.CategoryListResponse, error) {
	categories, err := s.repo.FindAll(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	resp := mappers.ToCategoryListResponse(categories)
	return &resp, nil
}

func (s *CategoryService) CountPostsByCategory() (*dtos.CategoryCountListResponse, error) {
	rows, err := s.repo.CountPostsByCategory()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch post counts by category: %w", err)
	}
	resp := mappers.ToCategoryCountListResponse(rows)
	return &resp, nil
}
