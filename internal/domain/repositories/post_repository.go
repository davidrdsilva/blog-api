package repositories

import (
	"github.com/davidrdsilva/blog-api/internal/domain/models"
)

// PostRepository defines the interface for post data access
// This is part of the domain layer and keeps business logic decoupled from infrastructure
type PostRepository interface {
	// Create inserts a new post into the database
	Create(post *models.Post) error

	// FindByID retrieves a post by its UUID
	FindByID(id string) (*models.Post, error)

	// FindAll retrieves posts with filtering, pagination, and sorting
	FindAll(filters models.PostFilters) ([]*models.Post, *models.PaginationMeta, error)

	// Update modifies an existing post
	Update(id string, post *models.Post) error

	// Delete removes a post by its UUID
	Delete(id string) error

	// Exists checks if a post with the given ID exists
	Exists(id string) (bool, error)
}
