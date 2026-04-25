package repository

import (
	"fmt"
	"strings"

	"github.com/davidrdsilva/blog-api/internal/domain/models"
	"github.com/davidrdsilva/blog-api/internal/domain/repositories"
	"gorm.io/gorm"
)

// PostgresTagRepository implements TagRepository using PostgreSQL
type PostgresTagRepository struct {
	db *gorm.DB
}

// NewPostgresTagRepository creates a new PostgreSQL tag repository
func NewPostgresTagRepository(db *gorm.DB) repositories.TagRepository {
	return &PostgresTagRepository{db: db}
}

func (r *PostgresTagRepository) FindAll(filters models.TagFilters) ([]*models.Tag, error) {
	var tags []*models.Tag
	query := r.db.Model(&models.Tag{})

	if search := strings.TrimSpace(filters.Search); search != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(search)+"%")
	}

	if err := query.Order("name ASC").Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch tags: %w", err)
	}
	return tags, nil
}

func (r *PostgresTagRepository) FindByNames(names []string) ([]*models.Tag, error) {
	normalized := normalizeTagNames(names)
	if len(normalized) == 0 {
		return nil, nil
	}

	var tags []*models.Tag
	if err := r.db.Where("LOWER(name) IN ?", normalized).Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch tags by names: %w", err)
	}
	return tags, nil
}

func (r *PostgresTagRepository) FindOrCreateByNames(names []string) ([]*models.Tag, error) {
	// Preserve the user's input order so the response keeps the order they typed.
	type entry struct {
		original   string
		normalized string
	}
	seen := make(map[string]struct{})
	entries := make([]entry, 0, len(names))
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
		entries = append(entries, entry{original: trimmed, normalized: key})
	}
	if len(entries) == 0 {
		return nil, nil
	}

	keys := make([]string, len(entries))
	for i, e := range entries {
		keys[i] = e.normalized
	}

	var existing []*models.Tag
	if err := r.db.Where("LOWER(name) IN ?", keys).Find(&existing).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch existing tags: %w", err)
	}

	byName := make(map[string]*models.Tag, len(existing))
	for _, t := range existing {
		byName[strings.ToLower(t.Name)] = t
	}

	result := make([]*models.Tag, 0, len(entries))
	toCreate := make([]*models.Tag, 0)
	for _, e := range entries {
		if t, ok := byName[e.normalized]; ok {
			result = append(result, t)
			continue
		}
		nt := &models.Tag{Name: e.original}
		toCreate = append(toCreate, nt)
		result = append(result, nt)
	}

	if len(toCreate) > 0 {
		// On unique-name conflict another concurrent request may have inserted
		// the same tag — fall back to a re-read for the conflicting names.
		if err := r.db.Create(&toCreate).Error; err != nil {
			// Re-query for everything; safer than parsing the driver error.
			var refreshed []*models.Tag
			if rerr := r.db.Where("LOWER(name) IN ?", keys).Find(&refreshed).Error; rerr != nil {
				return nil, fmt.Errorf("failed to create tags: %w", err)
			}
			byName = make(map[string]*models.Tag, len(refreshed))
			for _, t := range refreshed {
				byName[strings.ToLower(t.Name)] = t
			}
			result = result[:0]
			for _, e := range entries {
				if t, ok := byName[e.normalized]; ok {
					result = append(result, t)
				}
			}
			return result, nil
		}
	}

	return result, nil
}

func normalizeTagNames(names []string) []string {
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
