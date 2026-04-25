package dtos

// CategoryResponse represents a single category in API responses
type CategoryResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// CategoryListResponse wraps a list of categories under the standard data envelope.
type CategoryListResponse struct {
	Data []CategoryResponse `json:"data"`
}

// CategoryCountResponse is one row of the by-category count breakdown.
type CategoryCountResponse struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	TotalPosts int64  `json:"total_posts"`
}

// CategoryCountListResponse wraps the by-category count breakdown.
type CategoryCountListResponse struct {
	Data []CategoryCountResponse `json:"data"`
}
