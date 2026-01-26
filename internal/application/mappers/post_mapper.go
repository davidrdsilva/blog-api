package mappers

import (
	"time"

	"github.com/davidrdsilva/blog-api/internal/application/dtos"
	"github.com/davidrdsilva/blog-api/internal/domain/models"
)

// ToPostResponse converts a domain Post to a PostResponse DTO
func ToPostResponse(post *models.Post) dtos.PostResponse {
	return dtos.PostResponse{
		ID:          post.ID,
		Title:       post.Title,
		Subtitle:    post.Subtitle,
		Description: post.Description,
		Image:       post.Image,
		Date:        post.Date.Format(time.RFC3339),
		Author:      post.Author,
		Content:     post.Content,
		CreatedAt:   post.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   post.UpdatedAt.Format(time.RFC3339),
	}
}

// ToPostListResponse converts a slice of Posts to a PostListResponse
func ToPostListResponse(posts []*models.Post, meta *models.PaginationMeta) dtos.PostListResponse {
	responses := make([]dtos.PostResponse, len(posts))
	for i, post := range posts {
		responses[i] = ToPostResponse(post)
	}

	return dtos.PostListResponse{
		Data: responses,
		Meta: *meta,
	}
}

// CreatePostRequestToPost converts a CreatePostRequest to a domain Post
func CreatePostRequestToPost(req dtos.CreatePostRequest) *models.Post {
	return &models.Post{
		Title:       req.Title,
		Subtitle:    req.Subtitle,
		Description: req.Description,
		Image:       req.Image,
		Author:      req.Author,
		Content:     req.Content,
		Date:        time.Now(),
	}
}

// UpdatePostRequestToPost applies UpdatePostRequest fields to a Post
// Only updates fields that are provided (not nil)
func UpdatePostRequestToPost(post *models.Post, req dtos.UpdatePostRequest) {
	if req.Title != nil {
		post.Title = *req.Title
	}
	if req.Subtitle != nil {
		post.Subtitle = req.Subtitle
	}
	if req.Description != nil {
		post.Description = *req.Description
	}
	if req.Image != nil {
		post.Image = *req.Image
	}
	if req.Content != nil {
		post.Content = req.Content
	}
}
