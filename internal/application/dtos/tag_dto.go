package dtos

// TagResponse represents a single tag in API responses
type TagResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TagListResponse wraps a list of tags under the standard data envelope.
type TagListResponse struct {
	Data []TagResponse `json:"data"`
}
