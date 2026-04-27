package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Post represents a blog post entity
type Post struct {
	ID                     string           `gorm:"type:uuid;primaryKey" json:"id"`
	Title                  string           `gorm:"type:varchar(200);not null" json:"title"`
	Subtitle               *string          `gorm:"type:varchar(300)" json:"subtitle"`
	Description            string           `gorm:"type:varchar(100);not null" json:"description"`
	Image                  string           `gorm:"type:varchar(2048);not null" json:"image"`
	Date                   time.Time        `gorm:"type:timestamp with time zone;not null" json:"date"`
	Author                 string           `gorm:"type:varchar(100);not null" json:"author"`
	Content                *EditorJsContent `gorm:"type:jsonb" json:"content"`
	CategoryID             int              `gorm:"not null;index" json:"category_id"`
	Category               *Category        `gorm:"foreignKey:CategoryID;references:ID" json:"category,omitempty"`
	Tags                   []Tag            `gorm:"many2many:posts_tags;" json:"tags,omitempty"`
	Characters             []Character      `gorm:"many2many:posts_characters;" json:"characters,omitempty"`
	TotalViews             int              `gorm:"not null;default:0" json:"total_views"`
	WhitenestChapterNumber *int             `gorm:"uniqueIndex" json:"whitenest_chapter_number,omitempty"`
	Comments               []Comment        `gorm:"foreignKey:PostID;references:ID;constraint:OnDelete:CASCADE" json:"comments,omitempty"`
	CreatedAt              time.Time        `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt              time.Time        `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

// TableName specifies the table name for GORM
func (Post) TableName() string {
	return "posts"
}

// BeforeCreate is a GORM hook that runs before creating a new post
// Generates a UUID for the post if not already set
func (p *Post) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	if p.Date.IsZero() {
		p.Date = time.Now()
	}
	return nil
}

// PostFilters holds filtering options for querying posts
type PostFilters struct {
	Search     string
	Author     string
	CategoryID *int
	TagNames   []string
	// nil = no filter, true = only chapters, false = only non-chapters.
	IsWhitenestChapter *bool
	SortBy             string
	SortOrder          string
	Page               int
	Limit              int
	// Posts in categories flagged is_internal (e.g. "Drafts") are excluded from
	// public listings by default. OnlyInternalCategories returns *only* those
	// posts (used by the drafts endpoint); IncludeInternalCategories disables
	// the default exclusion (used by the admin morgue if needed).
	IncludeInternalCategories bool
	OnlyInternalCategories    bool
}

// PaginationMeta holds pagination metadata
type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"totalPages"`
	HasMore    bool  `json:"hasMore"`
}
