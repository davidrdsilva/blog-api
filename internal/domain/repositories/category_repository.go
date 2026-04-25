package repositories

import (
	"github.com/davidrdsilva/blog-api/internal/domain/models"
)

// CategoryRepository defines the interface for category data access
type CategoryRepository interface {
	// FindAll lists categories filtered by an optional case-insensitive name search.
	FindAll(filters models.CategoryFilters) ([]*models.Category, error)

	// FindByID retrieves a category by its integer ID.
	FindByID(id int) (*models.Category, error)

	// Case-insensitive. Returns (nil, nil) when no row matches.
	FindByName(name string) (*models.Category, error)

	// Exists returns whether a category with the given ID exists.
	Exists(id int) (bool, error)

	// CountPostsByCategory returns post counts grouped by category, including
	// categories with zero posts (so the UI can show empty buckets too).
	CountPostsByCategory() ([]models.CategoryWithCount, error)
}
