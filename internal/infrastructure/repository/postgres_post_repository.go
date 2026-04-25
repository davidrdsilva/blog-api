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
	err := r.db.
		Preload("Category").
		Preload("Tags").
		Where("id = ?", id).
		First(&post).Error
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

	// Apply category filter
	if filters.CategoryID != nil {
		query = query.Where("category_id = ?", *filters.CategoryID)
	}

	// Apply tag-name filter (OR semantics: posts that have ANY of the named tags).
	// We use a subquery (rather than JOIN) so the row count from the main query
	// stays correct even when a post matches multiple tags.
	if names := normalizeTagFilterNames(filters.TagNames); len(names) > 0 {
		query = query.Where(
			"id IN (?)",
			r.db.Table("posts_tags AS pt").
				Select("pt.post_id").
				Joins("JOIN tags AS t ON t.id = pt.tag_id").
				Where("LOWER(t.name) IN ?", names),
		)
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

	// Execute query (preload category + tags so the list view can show them).
	if err := query.Preload("Category").Preload("Tags").Find(&posts).Error; err != nil {
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

// IncrementViews atomically bumps total_views by 1 for the given post.
// Returns nil silently for unknown IDs; callers (the worker) shouldn't fail
// just because a post was deleted between the read request and the job run.
func (r *PostgresPostRepository) IncrementViews(id string) error {
	return r.db.Model(&models.Post{}).
		Where("id = ?", id).
		UpdateColumn("total_views", gorm.Expr("total_views + 1")).Error
}

// FindMostViewed returns the top-N posts by total_views.
func (r *PostgresPostRepository) FindMostViewed(limit int) ([]*models.Post, error) {
	if limit <= 0 {
		limit = 5
	}
	var posts []*models.Post
	err := r.db.
		Preload("Category").
		Preload("Tags").
		Order("total_views DESC").
		Order("date DESC").
		Limit(limit).
		Find(&posts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch most viewed posts: %w", err)
	}
	return posts, nil
}

// FindSimilar returns posts that share tags with the given source post,
// ranked by the count of shared tags (DESC) then by date (DESC). The source
// post is always excluded.
//
// Implemented as two queries: the first ranks candidate IDs in SQL, the
// second preloads Category + Tags. We keep them separate because GORM's
// preloads don't compose cleanly with a GROUP BY + custom SELECT.
func (r *PostgresPostRepository) FindSimilar(postID string, limit int) ([]*models.Post, error) {
	if limit <= 0 {
		limit = 5
	}

	type ranked struct {
		ID         string
		SharedTags int64
	}

	var rows []ranked
	sourceTagIDs := r.db.
		Table("posts_tags").
		Select("tag_id").
		Where("post_id = ?", postID)

	if err := r.db.
		Table("posts AS p").
		Select("p.id AS id, COUNT(pt.tag_id) AS shared_tags").
		Joins("JOIN posts_tags pt ON pt.post_id = p.id").
		Where("pt.tag_id IN (?) AND p.id != ?", sourceTagIDs, postID).
		Group("p.id, p.date").
		Order("shared_tags DESC, p.date DESC").
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to rank similar posts: %w", err)
	}

	if len(rows) == 0 {
		return nil, nil
	}

	ids := make([]string, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}

	var posts []*models.Post
	if err := r.db.
		Preload("Category").
		Preload("Tags").
		Where("id IN ?", ids).
		Find(&posts).Error; err != nil {
		return nil, fmt.Errorf("failed to load similar posts: %w", err)
	}

	// Re-order to match the SQL ranking — `IN ?` doesn't preserve order.
	byID := make(map[string]*models.Post, len(posts))
	for _, p := range posts {
		byID[p.ID] = p
	}
	ordered := make([]*models.Post, 0, len(rows))
	for _, row := range rows {
		if p, ok := byID[row.ID]; ok {
			ordered = append(ordered, p)
		}
	}
	return ordered, nil
}

// ReplaceTags resets the tag set associated with a post. Used by Update so the
// caller can supply a full replacement list of tags.
func (r *PostgresPostRepository) ReplaceTags(postID string, tags []*models.Tag) error {
	target := &models.Post{ID: postID}
	tagSlice := make([]models.Tag, len(tags))
	for i, t := range tags {
		tagSlice[i] = *t
	}
	if err := r.db.Model(target).Association("Tags").Replace(tagSlice); err != nil {
		return fmt.Errorf("failed to replace tags: %w", err)
	}
	return nil
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

// normalizeTagFilterNames lowercases and dedupes filter values, dropping empties.
func normalizeTagFilterNames(names []string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(names))
	for _, n := range names {
		trimmed := strings.TrimSpace(n)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	return out
}
