package services

import (
	"fmt"
	"strings"

	"github.com/davidrdsilva/blog-api/config"
	"github.com/davidrdsilva/blog-api/internal/application/dtos"
	"github.com/davidrdsilva/blog-api/internal/application/mappers"
	"github.com/davidrdsilva/blog-api/internal/domain/models"
	"github.com/davidrdsilva/blog-api/internal/domain/repositories"
	"github.com/google/uuid"
)

// PostService handles business logic for posts
type PostService struct {
	repo   repositories.PostRepository
	config *config.Config
}

// NewPostService creates a new post service
func NewPostService(repo repositories.PostRepository, cfg *config.Config) *PostService {
	return &PostService{
		repo:   repo,
		config: cfg,
	}
}

// CreatePost creates a new blog post
func (s *PostService) CreatePost(req dtos.CreatePostRequest) (*dtos.PostResponse, error) {
	// Validate image URL is from trusted storage
	if err := s.validateImageURL(req.Image); err != nil {
		return nil, fmt.Errorf("invalid image URL: %w", err)
	}

	// Convert DTO to domain model
	post := mappers.CreatePostRequestToPost(req)

	// Save to repository
	if err := s.repo.Create(post); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	// Convert back to response DTO
	response := mappers.ToPostResponse(post)
	return &response, nil
}

// GetPost retrieves a post by ID
func (s *PostService) GetPost(id string) (*dtos.PostResponse, error) {
	// Validate UUID format
	if !isValidUUID(id) {
		return nil, fmt.Errorf("invalid UUID format")
	}

	post, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch post: %w", err)
	}
	if post == nil {
		return nil, nil // Not found
	}

	response := mappers.ToPostResponse(post)
	return &response, nil
}

// ListPosts retrieves posts with pagination and filtering
func (s *PostService) ListPosts(filters models.PostFilters) (*dtos.PostListResponse, error) {
	posts, meta, err := s.repo.FindAll(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list posts: %w", err)
	}

	response := mappers.ToPostListResponse(posts, meta)
	return &response, nil
}

// UpdatePost updates an existing post
func (s *PostService) UpdatePost(id string, req dtos.UpdatePostRequest) (*dtos.PostResponse, error) {
	// Validate UUID format
	if !isValidUUID(id) {
		return nil, fmt.Errorf("invalid UUID format")
	}

	// Validate image URL if provided
	if req.Image != nil {
		if err := s.validateImageURL(*req.Image); err != nil {
			return nil, fmt.Errorf("invalid image URL: %w", err)
		}
	}

	// Fetch existing post
	post, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch post: %w", err)
	}
	if post == nil {
		return nil, nil // Not found
	}

	// Apply updates
	mappers.UpdatePostRequestToPost(post, req)

	// Save changes
	if err := s.repo.Update(id, post); err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	// Fetch updated post to get latest timestamps
	updatedPost, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated post: %w", err)
	}

	response := mappers.ToPostResponse(updatedPost)
	return &response, nil
}

// DeletePost deletes a post by ID
func (s *PostService) DeletePost(id string) error {
	// Validate UUID format
	if !isValidUUID(id) {
		return fmt.Errorf("invalid UUID format")
	}

	err := s.repo.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}
	return nil
}

// validateImageURL checks if the image URL is from the trusted storage domain
func (s *PostService) validateImageURL(imageURL string) error {
	// Get the public URL from MinIO config
	trustedDomain := s.config.MinIO.PublicURL

	// Check if the image URL starts with the trusted domain
	if !strings.HasPrefix(imageURL, trustedDomain) {
		return fmt.Errorf("image must be uploaded via /api/upload endpoint")
	}

	return nil
}

// isValidUUID checks if a string is a valid UUID
func isValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
