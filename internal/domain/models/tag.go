package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Tag represents a label that can be associated with one or more posts.
// Tags are user-supplied (created on the fly when a post is created), so we
// use a UUID id instead of an auto-increment to avoid coordination on inserts.
type Tag struct {
	ID   string `gorm:"type:uuid;primaryKey" json:"id"`
	Name string `gorm:"type:varchar(60);not null;uniqueIndex" json:"name"`
}

// TableName specifies the table name for GORM
func (Tag) TableName() string {
	return "tags"
}

// BeforeCreate generates a UUID for new tags.
func (t *Tag) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

// TagFilters holds filtering options for querying tags
type TagFilters struct {
	Search string
}

// PostsTag is the explicit join row between posts and tags. The spec calls for
// a UUID primary key on the join table itself, so we declare it explicitly
// rather than relying on GORM's default composite-key join.
type PostsTag struct {
	ID     string `gorm:"type:uuid;primaryKey" json:"id"`
	PostID string `gorm:"type:uuid;not null;index" json:"post_id"`
	TagID  string `gorm:"type:uuid;not null;index" json:"tag_id"`
}

// TableName specifies the table name for GORM
func (PostsTag) TableName() string {
	return "posts_tags"
}

// BeforeCreate generates a UUID for new join rows.
func (pt *PostsTag) BeforeCreate(tx *gorm.DB) error {
	if pt.ID == "" {
		pt.ID = uuid.New().String()
	}
	return nil
}
