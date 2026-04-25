package repositories

import (
	"github.com/davidrdsilva/blog-api/internal/domain/models"
)

// TagRepository defines the interface for tag data access
type TagRepository interface {
	// FindAll lists tags filtered by an optional case-insensitive name search.
	FindAll(filters models.TagFilters) ([]*models.Tag, error)

	// FindByNames returns all tags whose names match (case-insensitive) any
	// of the provided names. Used both for filtering posts by tag and for
	// the find-or-create flow when saving a post.
	FindByNames(names []string) ([]*models.Tag, error)

	// FindOrCreateByNames returns existing tags for the given names and
	// creates any that do not yet exist. Names are normalized (trimmed and
	// deduplicated case-insensitively) before lookup.
	FindOrCreateByNames(names []string) ([]*models.Tag, error)
}
