package repository

import (
	"fmt"
	"strings"

	"github.com/davidrdsilva/blog-api/internal/domain/models"
	"github.com/davidrdsilva/blog-api/internal/domain/repositories"
	"gorm.io/gorm"
)

// PostgresCategoryRepository implements CategoryRepository using PostgreSQL
type PostgresCategoryRepository struct {
	db *gorm.DB
}

// NewPostgresCategoryRepository creates a new PostgreSQL category repository
func NewPostgresCategoryRepository(db *gorm.DB) repositories.CategoryRepository {
	return &PostgresCategoryRepository{db: db}
}

func (r *PostgresCategoryRepository) FindAll(filters models.CategoryFilters) ([]*models.Category, error) {
	var categories []*models.Category
	query := r.db.Model(&models.Category{})

	if search := strings.TrimSpace(filters.Search); search != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(search)+"%")
	}

	if err := query.Order("name ASC").Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}
	return categories, nil
}

func (r *PostgresCategoryRepository) FindByID(id int) (*models.Category, error) {
	var category models.Category
	err := r.db.Where("id = ?", id).First(&category).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &category, err
}

func (r *PostgresCategoryRepository) FindByName(name string) (*models.Category, error) {
	var category models.Category
	err := r.db.Where("LOWER(name) = LOWER(?)", strings.TrimSpace(name)).First(&category).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &category, err
}

func (r *PostgresCategoryRepository) Exists(id int) (bool, error) {
	var count int64
	err := r.db.Model(&models.Category{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *PostgresCategoryRepository) CountPostsByCategory() ([]models.CategoryWithCount, error) {
	var rows []models.CategoryWithCount
	// LEFT JOIN so categories with zero posts still appear in the breakdown.
	err := r.db.Table("categories AS c").
		Select("c.id AS id, c.name AS name, COUNT(p.id) AS total_posts").
		Joins("LEFT JOIN posts AS p ON p.category_id = c.id").
		Group("c.id, c.name").
		Order("c.name ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count posts by category: %w", err)
	}
	return rows, nil
}
