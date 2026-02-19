package repository

import (
	"github.com/davidrdsilva/blog-api/internal/domain/models"
	"gorm.io/gorm"
)

type PostgresCommentRepository struct {
	db *gorm.DB
}

func NewPostgresCommentRepository(db *gorm.DB) *PostgresCommentRepository {
	return &PostgresCommentRepository{db: db}
}

func (r *PostgresCommentRepository) Create(comment *models.Comment) error {
	return r.db.Create(comment).Error
}

func (r *PostgresCommentRepository) FindByID(id string) (*models.Comment, error) {
	var comment models.Comment
	if err := r.db.First(&comment, id).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *PostgresCommentRepository) FindAllByPostID(postID string) ([]*models.Comment, error) {
	var comments []*models.Comment
	if err := r.db.Where("post_id = ?", postID).Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *PostgresCommentRepository) FindAll(filters models.CommentFilters) ([]*models.Comment, error) {
	var comments []*models.Comment
	if err := r.db.Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *PostgresCommentRepository) Update(id string, comment *models.Comment) error {
	return r.db.Save(comment).Error
}

func (r *PostgresCommentRepository) Delete(id string) error {
	return r.db.Delete(&models.Comment{}, id).Error
}

func (r *PostgresCommentRepository) Exists(id string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Comment{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}
