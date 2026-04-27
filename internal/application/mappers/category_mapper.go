package mappers

import (
	"github.com/davidrdsilva/blog-api/internal/application/dtos"
	"github.com/davidrdsilva/blog-api/internal/domain/models"
)

// ToCategoryResponse converts a domain Category to a CategoryResponse DTO
func ToCategoryResponse(c *models.Category) dtos.CategoryResponse {
	return dtos.CategoryResponse{
		ID:         c.ID,
		Name:       c.Name,
		IsInternal: c.IsInternal,
	}
}

// ToCategoryListResponse converts a slice of domain Categories to the list DTO
func ToCategoryListResponse(categories []*models.Category) dtos.CategoryListResponse {
	out := make([]dtos.CategoryResponse, len(categories))
	for i, c := range categories {
		out[i] = ToCategoryResponse(c)
	}
	return dtos.CategoryListResponse{Data: out}
}

// ToCategoryCountListResponse converts the count-by-category projection to its DTO
func ToCategoryCountListResponse(rows []models.CategoryWithCount) dtos.CategoryCountListResponse {
	out := make([]dtos.CategoryCountResponse, len(rows))
	for i, r := range rows {
		out[i] = dtos.CategoryCountResponse{
			ID:         r.ID,
			Name:       r.Name,
			TotalPosts: r.TotalPosts,
		}
	}
	return dtos.CategoryCountListResponse{Data: out}
}
