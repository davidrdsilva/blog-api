package services

import (
	"fmt"
	"strings"

	"github.com/davidrdsilva/blog-api/config"
	"github.com/davidrdsilva/blog-api/internal/application/dtos"
	"github.com/davidrdsilva/blog-api/internal/application/jobs"
	"github.com/davidrdsilva/blog-api/internal/application/mappers"
	"github.com/davidrdsilva/blog-api/internal/domain/models"
	"github.com/davidrdsilva/blog-api/internal/domain/repositories"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
	"github.com/google/uuid"
)

// PostService handles business logic for posts
type PostService struct {
	repo         repositories.PostRepository
	categoryRepo repositories.CategoryRepository
	tagRepo      repositories.TagRepository
	config       *config.Config
	jobCh        chan<- jobs.GenerateCommentsJob
	viewCh       chan<- jobs.IncrementPostViewsJob
	logger       *logging.Logger
}

// NewPostService creates a new post service
func NewPostService(
	repo repositories.PostRepository,
	categoryRepo repositories.CategoryRepository,
	tagRepo repositories.TagRepository,
	cfg *config.Config,
	jobCh chan<- jobs.GenerateCommentsJob,
	viewCh chan<- jobs.IncrementPostViewsJob,
	logger *logging.Logger,
) *PostService {
	return &PostService{
		repo:         repo,
		categoryRepo: categoryRepo,
		tagRepo:      tagRepo,
		config:       cfg,
		jobCh:        jobCh,
		viewCh:       viewCh,
		logger:       logger,
	}
}

func (s *PostService) CreatePost(req dtos.CreatePostRequest) (*dtos.PostResponse, error) {
	if err := s.validateImageURL(req.Image); err != nil {
		return nil, fmt.Errorf("invalid image URL: %w", err)
	}

	// Reject category IDs that don't resolve to a real row before persisting.
	exists, err := s.categoryRepo.Exists(req.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify category: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("invalid category: category %d does not exist", req.CategoryID)
	}

	// Convert DTO to domain model
	post := mappers.CreatePostRequestToPost(req)

	// Resolve tag names to existing-or-new tag rows. We do this *before* the
	// post insert so the join rows can be written in the same Create call —
	// GORM only inserts the join when the related entities have IDs.
	if len(req.Tags) > 0 {
		tags, terr := s.tagRepo.FindOrCreateByNames(req.Tags)
		if terr != nil {
			return nil, fmt.Errorf("failed to resolve tags: %w", terr)
		}
		tagSlice := make([]models.Tag, len(tags))
		for i, t := range tags {
			tagSlice[i] = *t
		}
		post.Tags = tagSlice
	}

	if err := s.repo.Create(post); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	// Re-fetch so Category is populated for the response.
	saved, err := s.repo.FindByID(post.ID)
	if err != nil || saved == nil {
		// Fall back to the in-memory post; not fatal.
		saved = post
	}

	// Dispatch AI comment generation asynchronously (non-blocking)
	if s.jobCh != nil {
		select {
		case s.jobCh <- jobs.GenerateCommentsJob{
			PostID:  saved.ID,
			Title:   saved.Title,
			Content: saved.Content,
		}:
			s.logger.Debug("AI comment job created", logging.F("postId", saved.ID))
		default:
			// Worker is behind and the buffer is full — drop the job rather than
			// blocking the HTTP response. The post is already saved successfully.
			s.logger.Warn("AI comment job dropped: job channel full", logging.F("postId", saved.ID))
		}
	}

	response := mappers.ToPostResponse(saved)
	return &response, nil
}

func (s *PostService) GetPost(id string) (*dtos.PostResponse, error) {
	if !isValidUUID(id) {
		return nil, fmt.Errorf("invalid UUID format")
	}

	post, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch post: %w", err)
	}
	if post == nil {
		return nil, nil
	}

	// Fire-and-forget: bump total_views asynchronously so the read path stays
	// fast and stays decoupled from a write that can fail independently. If
	// the buffer is full we drop the increment rather than blocking.
	if s.viewCh != nil {
		select {
		case s.viewCh <- jobs.IncrementPostViewsJob{PostID: post.ID}:
		default:
			s.logger.Warn("view increment dropped: channel full",
				logging.F("postId", post.ID),
			)
		}
	}

	response := mappers.ToPostResponse(post)
	return &response, nil
}

// GetMostViewed returns the top-N posts by total_views without pagination.
func (s *PostService) GetMostViewed(limit int) (*dtos.PostListResponse, error) {
	posts, err := s.repo.FindMostViewed(limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch most viewed posts: %w", err)
	}
	// Most-viewed has no pagination, so we surface a synthetic meta with the
	// length so the response shape is consistent with the regular list.
	meta := &models.PaginationMeta{
		Total:      int64(len(posts)),
		Page:       1,
		Limit:      limit,
		TotalPages: 1,
		HasMore:    false,
	}
	response := mappers.ToPostListResponse(posts, meta)
	return &response, nil
}

func (s *PostService) ListPosts(filters models.PostFilters) (*dtos.PostListResponse, error) {
	posts, meta, err := s.repo.FindAll(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list posts: %w", err)
	}

	response := mappers.ToPostListResponse(posts, meta)
	return &response, nil
}

func (s *PostService) UpdatePost(id string, req dtos.UpdatePostRequest) (*dtos.PostResponse, error) {
	if !isValidUUID(id) {
		return nil, fmt.Errorf("invalid UUID format")
	}

	if req.Image != nil {
		if err := s.validateImageURL(*req.Image); err != nil {
			return nil, fmt.Errorf("invalid image URL: %w", err)
		}
	}

	if req.CategoryID != nil {
		exists, err := s.categoryRepo.Exists(*req.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to verify category: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("invalid category: category %d does not exist", *req.CategoryID)
		}
	}

	post, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch post: %w", err)
	}
	if post == nil {
		return nil, nil
	}

	// Apply updates
	mappers.UpdatePostRequestToPost(post, req)

	// Save changes
	if err := s.repo.Update(id, post); err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	// If the request includes tags, treat it as a full replacement of the set.
	if req.Tags != nil {
		tags, terr := s.tagRepo.FindOrCreateByNames(*req.Tags)
		if terr != nil {
			return nil, fmt.Errorf("failed to resolve tags: %w", terr)
		}
		if err := s.repo.ReplaceTags(id, tags); err != nil {
			return nil, fmt.Errorf("failed to replace tags: %w", err)
		}
	}

	updatedPost, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated post: %w", err)
	}

	// Only regenerate comments when the post body itself changed.
	if req.Content != nil && s.jobCh != nil {
		select {
		case s.jobCh <- jobs.GenerateCommentsJob{
			PostID:  updatedPost.ID,
			Title:   updatedPost.Title,
			Content: updatedPost.Content,
		}:
			s.logger.Debug("AI comment job created for updated post", logging.F("postId", updatedPost.ID))
		default:
			s.logger.Warn("AI comment job dropped: job channel full", logging.F("postId", updatedPost.ID))
		}
	}

	response := mappers.ToPostResponse(updatedPost)
	return &response, nil
}

// DeletePost deletes a post by ID
func (s *PostService) DeletePost(id string) error {
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
	trustedDomain := s.config.MinIO.PublicURL

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
