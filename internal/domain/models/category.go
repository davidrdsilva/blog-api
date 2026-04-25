package models

// Category represents a post category. Each post belongs to exactly one
// category. Categories use auto-incrementing integer IDs because they are a
// small, curated set of labels (unlike tags, which are user-supplied and use
// UUIDs).
type Category struct {
	ID   int    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"type:varchar(60);not null;uniqueIndex" json:"name"`
}

// TableName specifies the table name for GORM
func (Category) TableName() string {
	return "categories"
}

// CategoryFilters holds filtering options for querying categories
type CategoryFilters struct {
	Search string
}

// CategoryWithCount is a projection used by the by-category counter endpoint.
type CategoryWithCount struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	TotalPosts int64  `json:"total_posts"`
}
