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

	// ReplaceTags fully replaces the set of tags associated with a post.
	// Used by updates that include a tags array.
	ReplaceTags(postID string, tags []*models.Tag) error

	// IncrementViews adds 1 to total_views atomically. Called from the view
	// counter worker, decoupled from the read path.
	IncrementViews(id string) error

	// FindMostViewed returns up to `limit` posts ordered by total_views DESC.
	// No pagination — the response is intentionally small.
	FindMostViewed(limit int) ([]*models.Post, error)

	// FindSimilar returns up to `limit` posts that share tags with the given
	// post, ranked by the count of overlapping tags then by date DESC. The
	// source post is always excluded. If the source has no tags, returns an
	// empty slice.
	FindSimilar(postID string, limit int) ([]*models.Post, error)

	// Returns (nil, nil) when no chapter has that number.
	FindWhitenestChapterByNumber(number int) (*models.Post, error)

	// Either side may be nil at the extremes of the series.
	FindAdjacentWhitenestChapters(number int) (previous, next *models.Post, err error)

	// Returns 0 when no chapters exist yet.
	MaxWhitenestChapterNumber() (int, error)
}
