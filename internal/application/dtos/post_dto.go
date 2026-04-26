package dtos

import (
	"time"

	"github.com/davidrdsilva/blog-api/internal/domain/models"
)

// CreatePostRequest represents the request body for creating a post
type CreatePostRequest struct {
	Title       string                  `json:"title" binding:"required,min=1,max=200"`
	Subtitle    *string                 `json:"subtitle" binding:"omitempty,max=300"`
	Description string                  `json:"description" binding:"required,min=1,max=100"`
	Image       string                  `json:"image" binding:"required,url"`
	Author      string                  `json:"author" binding:"required,min=1,max=100"`
	Content     *models.EditorJsContent `json:"content"`
	Date        *time.Time              `json:"date" binding:"omitempty"`
	CategoryID  int                     `json:"category_id" binding:"required,min=1"`
	Tags        []string                `json:"tags" binding:"omitempty,dive,min=1,max=60"`
	// Auto-assigned to MAX+1 when omitted on a Whitenest post; rejected on others.
	WhitenestChapterNumber *int `json:"whitenest_chapter_number,omitempty" binding:"omitempty,min=1"`
}

// UpdatePostRequest represents the request body for updating a post
type UpdatePostRequest struct {
	Title       *string                 `json:"title" binding:"omitempty,min=1,max=200"`
	Subtitle    *string                 `json:"subtitle" binding:"omitempty,max=300"`
	Description *string                 `json:"description" binding:"omitempty,min=1,max=100"`
	Image       *string                 `json:"image" binding:"omitempty,url"`
	Content     *models.EditorJsContent `json:"content"`
	Date        *time.Time              `json:"date" binding:"omitempty"`
	CategoryID  *int                    `json:"category_id" binding:"omitempty,min=1"`
	// Tags is treated as a full replacement when present (nil means "leave as is").
	Tags                   *[]string `json:"tags" binding:"omitempty,dive,min=1,max=60"`
	WhitenestChapterNumber *int      `json:"whitenest_chapter_number,omitempty" binding:"omitempty,min=1"`
}

// PostResponse represents a single post in API responses
type PostResponse struct {
	ID                     string                  `json:"id"`
	Title                  string                  `json:"title"`
	Subtitle               *string                 `json:"subtitle"`
	Description            string                  `json:"description"`
	Image                  string                  `json:"image"`
	Date                   string                  `json:"date"`
	Author                 string                  `json:"author"`
	Content                *models.EditorJsContent `json:"content"`
	CategoryID             int                     `json:"category_id"`
	Category               *CategoryResponse       `json:"category,omitempty"`
	Tags                   []TagResponse           `json:"tags"`
	TotalViews             int                     `json:"total_views"`
	WhitenestChapterNumber *int                    `json:"whitenest_chapter_number,omitempty"`
	CreatedAt              string                  `json:"createdAt"`
	UpdatedAt              string                  `json:"updatedAt"`
}

type WhitenestChapterRef struct {
	ID                     string `json:"id"`
	Title                  string `json:"title"`
	WhitenestChapterNumber int    `json:"whitenest_chapter_number"`
}

// WhitenestChapterSummary is a lightweight chapter representation for list
// views (e.g. the chapters sidebar) — image, title, tags, and chapter number.
type WhitenestChapterSummary struct {
	ID                     string        `json:"id"`
	Title                  string        `json:"title"`
	Image                  string        `json:"image"`
	Tags                   []TagResponse `json:"tags"`
	WhitenestChapterNumber int           `json:"whitenest_chapter_number"`
}

type WhitenestChapterResponse struct {
	Chapter  PostResponse         `json:"chapter"`
	Previous *WhitenestChapterRef `json:"previous"`
	Next     *WhitenestChapterRef `json:"next"`
}

// PostListResponse represents a paginated list of posts
type PostListResponse struct {
	Data []PostResponse        `json:"data"`
	Meta models.PaginationMeta `json:"meta"`
}

// SuccessResponse is a generic success response wrapper
type SuccessResponse struct {
	Data interface{} `json:"data"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    string              `json:"code"`
	Message string              `json:"message"`
	Details map[string][]string `json:"details,omitempty"`
}

// EditorJsUploadResponse represents the response format for Editor.js Image Tool
type EditorJsUploadResponse struct {
	Success int                  `json:"success"`
	File    *EditorJsFileInfo    `json:"file,omitempty"`
	Error   *EditorJsErrorDetail `json:"error,omitempty"`
}

// EditorJsFileInfo contains uploaded file information
type EditorJsFileInfo struct {
	URL string `json:"url"`
}

// EditorJsErrorDetail contains error information for Editor.js
type EditorJsErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// EditorJsURLResponse represents the response format for Editor.js Link Tool
type EditorJsURLResponse struct {
	Success int                  `json:"success"`
	Link    string               `json:"link,omitempty"`
	Meta    *URLMetadata         `json:"meta,omitempty"`
	Error   *EditorJsErrorDetail `json:"error,omitempty"`
}

// URLMetadata contains metadata extracted from a URL
type URLMetadata struct {
	Title       string        `json:"title,omitempty"`
	Description string        `json:"description,omitempty"`
	Image       *URLImageInfo `json:"image,omitempty"`
}

// URLImageInfo contains image URL information
type URLImageInfo struct {
	URL string `json:"url"`
}
