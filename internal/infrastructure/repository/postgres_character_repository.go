package repository

import (
	"fmt"
	"strings"

	"github.com/davidrdsilva/blog-api/internal/domain/models"
	"github.com/davidrdsilva/blog-api/internal/domain/repositories"
	"gorm.io/gorm"
)

type PostgresCharacterRepository struct {
	db *gorm.DB
}

func NewPostgresCharacterRepository(db *gorm.DB) repositories.CharacterRepository {
	return &PostgresCharacterRepository{db: db}
}

func (r *PostgresCharacterRepository) Create(character *models.Character) error {
	if err := r.db.Create(character).Error; err != nil {
		return fmt.Errorf("failed to create character: %w", err)
	}
	return nil
}

func (r *PostgresCharacterRepository) Update(id string, character *models.Character) error {
	res := r.db.Model(&models.Character{}).Where("id = ?", id).Updates(map[string]interface{}{
		"full_name":   character.FullName,
		"short_name":  character.ShortName,
		"description": character.Description,
		"occupation":  character.Occupation,
		"location":    character.Location,
		"portrait":    character.Portrait,
		"skills":      character.Skills,
	})
	if res.Error != nil {
		return fmt.Errorf("failed to update character: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *PostgresCharacterRepository) Delete(id string) error {
	res := r.db.Delete(&models.Character{}, "id = ?", id)
	if res.Error != nil {
		return fmt.Errorf("failed to delete character: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *PostgresCharacterRepository) FindByID(id string) (*models.Character, error) {
	var character models.Character
	err := r.db.Where("id = ?", id).First(&character).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch character: %w", err)
	}
	return &character, nil
}

func (r *PostgresCharacterRepository) FindAll(filters models.CharacterFilters) ([]*models.Character, error) {
	var characters []*models.Character
	query := r.db.Model(&models.Character{})

	if search := strings.TrimSpace(filters.Search); search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(full_name) LIKE ? OR LOWER(short_name) LIKE ?", like, like)
	}

	if err := query.Order("short_name ASC").Find(&characters).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch characters: %w", err)
	}
	return characters, nil
}

func (r *PostgresCharacterRepository) FindByIDs(ids []string) ([]*models.Character, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	var rows []*models.Character
	if err := r.db.Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch characters by ids: %w", err)
	}

	byID := make(map[string]*models.Character, len(rows))
	for _, c := range rows {
		byID[c.ID] = c
	}

	ordered := make([]*models.Character, 0, len(ids))
	for _, id := range ids {
		if c, ok := byID[id]; ok {
			ordered = append(ordered, c)
		}
	}
	return ordered, nil
}

func (r *PostgresCharacterRepository) Exists(id string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Character{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check character existence: %w", err)
	}
	return count > 0, nil
}
