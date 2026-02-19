package repositories

import "github.com/davidrdsilva/blog-api/internal/domain/models"

type CommentRepository interface {
	Create(comment *models.Comment) error
	FindByID(id string) (*models.Comment, error)
	FindAll(filters models.CommentFilters) ([]*models.Comment, error)
	Update(id string, comment *models.Comment) error
	Delete(id string) error
	FindAllByPostID(postID string) ([]*models.Comment, error)
	Exists(id string) (bool, error)
}
