package mappers

import (
	"time"

	"github.com/davidrdsilva/blog-api/internal/application/dtos"
	"github.com/davidrdsilva/blog-api/internal/domain/models"
)

func ToCharacterResponse(c *models.Character) dtos.CharacterResponse {
	return dtos.CharacterResponse{
		ID:          c.ID,
		FullName:    c.FullName,
		ShortName:   c.ShortName,
		Description: c.Description,
		Occupation:  c.Occupation,
		Location:    c.Location,
		Portrait:    c.Portrait,
		Skills:      c.Skills,
		CreatedAt:   c.CreatedAt.In(brt).Format(time.RFC3339),
		UpdatedAt:   c.UpdatedAt.In(brt).Format(time.RFC3339),
	}
}

func ToCharacterResponses(characters []models.Character) []dtos.CharacterResponse {
	out := make([]dtos.CharacterResponse, len(characters))
	for i := range characters {
		out[i] = ToCharacterResponse(&characters[i])
	}
	return out
}
