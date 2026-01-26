package repository

import (
	"fmt"
	"math"
	"strings"

	"github.com/davidrdsilva/blog-api/internal/domain/models"
	"github.com/davidrdsilva/blog-api/internal/domain/repositories"
	"gorm.io/gorm"
)

// PostgresPostRepository implements PostRepository using PostgreSQL
type PostgresPostRepository struct {
	db *gorm.DB
}

// NewPostgresPostRepository creates a new PostgreSQL post repository
func NewPostgresPostRepository(db *gorm.DB) repositories.PostRepository {
	return &PostgresPostRepository{db: db}
}

// Create inserts a new post into the database
func (r *PostgresPostRepository) Create(post *models.Post) error {
	return r.db.Create(post).Error
}

// FindByID retrieves a post by its UUID
func (r *PostgresPostRepository) FindByID(id string) (*models.Post, error) {
	var post models.Post
	err := r.db.Where("id = ?", id).First(&post).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &post, err
}

// FindAll retrieves posts with filtering, pagination, and sorting
func (r *PostgresPostRepository) FindAll(filters models.PostFilters) ([]*models.Post, *models.PaginationMeta, error) {
	var posts []*models.Post
	var total int64

	// Build query
	query := r.db.Model(&models.Post{})

	// Apply search filter using full-text search
	if filters.Search != "" {
		searchTerms := strings.TrimSpace(filters.Search)
		query = query.Where(
			"to_tsvector('english', title || ' ' || COALESCE(subtitle, '') || ' ' || description) @@ plainto_tsquery('english', ?)",
			searchTerms,
		)
	}

	// Apply author filter
	if filters.Author != "" {
		query = query.Where("author = ?", filters.Author)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to count posts: %w", err)
	}

	// Apply sorting
	sortBy := filters.SortBy
	if sortBy == "" {
		sortBy = "date"
	}
	sortOrder := filters.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}

	// Validate and sanitize sort fields to prevent SQL injection
	allowedSortFields := map[string]bool{
		"date":      true,
		"title":     true,
		"createdAt": true,
		"updatedAt": true,
	}
	if !allowedSortFields[sortBy] {
		sortBy = "date"
	}

	// Convert camelCase to snake_case for database column names
	dbSortBy := camelToSnake(sortBy)

	if strings.ToLower(sortOrder) != "asc" {
		sortOrder = "desc"
	}

	query = query.Order(fmt.Sprintf("%s %s", dbSortBy, sortOrder))

	// Apply pagination
	page := filters.Page
	if page < 1 {
		page = 1
	}
	limit := filters.Limit
	if limit < 1 {
		limit = 6
	}
	if limit > 50 {
		limit = 50
	}

	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	// Execute query
	if err := query.Find(&posts).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to fetch posts: %w", err)
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	hasMore := page < totalPages

	meta := &models.PaginationMeta{
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
		HasMore:    hasMore,
	}

	return posts, meta, nil
}

// Update modifies an existing post
func (r *PostgresPostRepository) Update(id string, post *models.Post) error {
	result := r.db.Model(&models.Post{}).Where("id = ?", id).Updates(post)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete removes a post by its UUID
func (r *PostgresPostRepository) Delete(id string) error {
	result := r.db.Where("id = ?", id).Delete(&models.Post{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Exists checks if a post with the given ID exists
func (r *PostgresPostRepository) Exists(id string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Post{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// camelToSnake converts camelCase to snake_case
func camelToSnake(s string) string {
	var result strings.Builder
	for i, char := range s {
		if i > 0 && char >= 'A' && char <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(char)
	}
	return strings.ToLower(result.String())
}
