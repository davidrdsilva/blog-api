package mappers

import (
	"github.com/davidrdsilva/blog-api/internal/application/dtos"
	"github.com/davidrdsilva/blog-api/internal/domain/models"
)

// ToTagResponse converts a domain Tag to a TagResponse DTO
func ToTagResponse(t *models.Tag) dtos.TagResponse {
	return dtos.TagResponse{
		ID:   t.ID,
		Name: t.Name,
	}
}

// ToTagListResponse converts a slice of domain Tags to the list DTO
func ToTagListResponse(tags []*models.Tag) dtos.TagListResponse {
	out := make([]dtos.TagResponse, len(tags))
	for i, t := range tags {
		out[i] = ToTagResponse(t)
	}
	return dtos.TagListResponse{Data: out}
}

// ToTagResponses converts a slice of (non-pointer) domain Tags into response DTOs.
// Used when reading the Tags relation off a Post (gorm returns []Tag, not []*Tag).
func ToTagResponses(tags []models.Tag) []dtos.TagResponse {
	out := make([]dtos.TagResponse, len(tags))
	for i := range tags {
		out[i] = ToTagResponse(&tags[i])
	}
	return out
}
