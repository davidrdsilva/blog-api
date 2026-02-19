package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Comment struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	PostID    string    `gorm:"type:uuid;not null" json:"postId"`
	Author    string    `gorm:"type:varchar(100);not null" json:"author"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP" json:"createdAt"`
}

func (Comment) TableName() string {
	return "comments"
}

func (c *Comment) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	return nil
}

type CommentFilters struct {
	PostID    string
	Author    string
	SortBy    string
	SortOrder string
	Page      int
	Limit     int
}
