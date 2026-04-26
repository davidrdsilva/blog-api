package repositories

import (
	"github.com/davidrdsilva/blog-api/internal/domain/models"
)

// CharacterRepository defines the interface for character data access.
type CharacterRepository interface {
	Create(character *models.Character) error
	Update(id string, character *models.Character) error
	Delete(id string) error
	FindByID(id string) (*models.Character, error)
	FindAll(filters models.CharacterFilters) ([]*models.Character, error)
	FindByIDs(ids []string) ([]*models.Character, error)
	Exists(id string) (bool, error)
}
