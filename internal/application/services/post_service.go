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
	"github.com/davidrdsilva/blog-api/internal/infrastructure/database"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
	"github.com/google/uuid"
)

// Matched as a substring by the post handler to map to WHITENEST_INVARIANT_VIOLATION.
const errWhitenestMismatch = "whitenest invariant: chapter number requires Whitenest category"

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

	cat, err := s.categoryRepo.FindByID(req.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify category: %w", err)
	}
	if cat == nil {
		return nil, fmt.Errorf("invalid category: category %d does not exist", req.CategoryID)
	}
	isWhitenestCategory := strings.EqualFold(cat.Name, database.WhitenestCategoryName)

	if req.WhitenestChapterNumber != nil && !isWhitenestCategory {
		return nil, fmt.Errorf("%s: provided number=%d on category=%q",
			errWhitenestMismatch, *req.WhitenestChapterNumber, cat.Name)
	}

	if isWhitenestCategory && req.WhitenestChapterNumber == nil {
		max, err := s.repo.MaxWhitenestChapterNumber()
		if err != nil {
			return nil, fmt.Errorf("failed to assign next chapter number: %w", err)
		}
		next := max + 1
		req.WhitenestChapterNumber = &next
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

	if saved.WhitenestChapterNumber == nil {
		s.dispatchAICommentJob(saved)
	}

	response := mappers.ToPostResponse(saved)
	return &response, nil
}

func (s *PostService) dispatchAICommentJob(post *models.Post) {
	if s.jobCh == nil {
		return
	}
	select {
	case s.jobCh <- jobs.GenerateCommentsJob{
		PostID:  post.ID,
		Title:   post.Title,
		Content: post.Content,
	}:
		s.logger.Debug("AI comment job created", logging.F("postId", post.ID))
	default:
		s.logger.Warn("AI comment job dropped: job channel full", logging.F("postId", post.ID))
	}
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

// GetSimilarPosts returns posts ranked by shared tags with the given post.
// Empty result if the source has no tags or no other tagged posts overlap.
func (s *PostService) GetSimilarPosts(id string, limit int) (*dtos.PostListResponse, error) {
	if !isValidUUID(id) {
		return nil, fmt.Errorf("invalid UUID format")
	}

	posts, err := s.repo.FindSimilar(id, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch similar posts: %w", err)
	}

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

	post, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch post: %w", err)
	}
	if post == nil {
		return nil, nil
	}

	// Re-validate the invariant against the post-update state.
	effectiveCategoryID := post.CategoryID
	if req.CategoryID != nil {
		effectiveCategoryID = *req.CategoryID
	}
	cat, err := s.categoryRepo.FindByID(effectiveCategoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify category: %w", err)
	}
	if cat == nil {
		return nil, fmt.Errorf("invalid category: category %d does not exist", effectiveCategoryID)
	}
	isWhitenestCategory := strings.EqualFold(cat.Name, database.WhitenestCategoryName)

	effectiveChapterNumber := post.WhitenestChapterNumber
	if req.WhitenestChapterNumber != nil {
		effectiveChapterNumber = req.WhitenestChapterNumber
	}

	if effectiveChapterNumber != nil && !isWhitenestCategory {
		return nil, fmt.Errorf("%s: post would have number=%d on category=%q",
			errWhitenestMismatch, *effectiveChapterNumber, cat.Name)
	}
	if isWhitenestCategory && effectiveChapterNumber == nil {
		max, err := s.repo.MaxWhitenestChapterNumber()
		if err != nil {
			return nil, fmt.Errorf("failed to assign next chapter number: %w", err)
		}
		next := max + 1
		req.WhitenestChapterNumber = &next
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

	if req.Content != nil && updatedPost.WhitenestChapterNumber == nil {
		s.dispatchAICommentJob(updatedPost)
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
