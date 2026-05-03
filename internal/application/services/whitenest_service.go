package services

import (
	"fmt"

	"github.com/davidrdsilva/blog-api/internal/application/dtos"
	"github.com/davidrdsilva/blog-api/internal/application/jobs"
	"github.com/davidrdsilva/blog-api/internal/application/mappers"
	"github.com/davidrdsilva/blog-api/internal/domain/models"
	"github.com/davidrdsilva/blog-api/internal/domain/repositories"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
)

// errChapterSetMismatch is matched as a substring by the whitenest handler to
// surface a 409 — the request's chapter set doesn't match what's currently in
// the DB (a publish/unpublish landed between page load and submit). Client is
// expected to re-fetch the chapter list and let the user redo the reorder.
const errChapterSetMismatch = "chapter set mismatch"

type WhitenestService struct {
	postRepo repositories.PostRepository
	viewCh   chan<- jobs.IncrementPostViewsJob
	logger   *logging.Logger
}

func NewWhitenestService(
	postRepo repositories.PostRepository,
	viewCh chan<- jobs.IncrementPostViewsJob,
	logger *logging.Logger,
) *WhitenestService {
	return &WhitenestService{
		postRepo: postRepo,
		viewCh:   viewCh,
		logger:   logger,
	}
}

// Returns (nil, nil) when no chapter has that number.
func (s *WhitenestService) GetChapterByNumber(number int) (*dtos.WhitenestChapterResponse, error) {
	if number < 1 {
		return nil, fmt.Errorf("invalid chapter number: %d", number)
	}

	post, err := s.postRepo.FindWhitenestChapterByNumber(number)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch chapter: %w", err)
	}
	if post == nil {
		return nil, nil
	}

	previous, next, err := s.postRepo.FindAdjacentWhitenestChapters(number)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch adjacent chapters: %w", err)
	}

	if s.viewCh != nil {
		select {
		case s.viewCh <- jobs.IncrementPostViewsJob{PostID: post.ID}:
		default:
			s.logger.Warn("view increment dropped: channel full",
				logging.F("postId", post.ID),
			)
		}
	}

	return &dtos.WhitenestChapterResponse{
		Chapter:  mappers.ToPostResponse(post),
		Previous: mappers.ToWhitenestChapterRef(previous),
		Next:     mappers.ToWhitenestChapterRef(next),
		Cast:     mappers.ToCharacterResponses(post.Characters),
	}, nil
}

// ListChapters returns every Whitenest chapter ordered by chapter number ASC.
func (s *WhitenestService) ListChapters() ([]dtos.WhitenestChapterSummary, error) {
	posts, err := s.postRepo.ListWhitenestChapters()
	if err != nil {
		return nil, fmt.Errorf("failed to list chapters: %w", err)
	}
	return mappers.ToWhitenestChapterSummaries(posts), nil
}

// ReorderChapters validates the request against the current chapter set and,
// if it covers every chapter exactly once with contiguous numbers 1..N,
// rewrites the numbers atomically. Mismatches return errChapterSetMismatch so
// the client can refresh and try again — the most common cause is a concurrent
// publish or unpublish between when the user opened the reorder UI and when
// they submitted it.
func (s *WhitenestService) ReorderChapters(req dtos.ReorderChaptersRequest) error {
	if err := validateReorder(req.Order); err != nil {
		return err
	}

	current, err := s.postRepo.ListWhitenestChapters()
	if err != nil {
		return fmt.Errorf("failed to load current chapters: %w", err)
	}
	if len(current) != len(req.Order) {
		return fmt.Errorf("%s: expected %d chapters, got %d",
			errChapterSetMismatch, len(current), len(req.Order))
	}

	currentIDs := make(map[string]struct{}, len(current))
	for _, p := range current {
		currentIDs[p.ID] = struct{}{}
	}
	seen := make(map[string]struct{}, len(req.Order))
	items := make([]models.ChapterOrderItem, len(req.Order))
	for i, item := range req.Order {
		if _, ok := currentIDs[item.PostID]; !ok {
			return fmt.Errorf("%s: post %s is not a Whitenest chapter",
				errChapterSetMismatch, item.PostID)
		}
		if _, dup := seen[item.PostID]; dup {
			return fmt.Errorf("duplicate post_id in order: %s", item.PostID)
		}
		seen[item.PostID] = struct{}{}
		items[i] = models.ChapterOrderItem{PostID: item.PostID, Number: item.Number}
	}

	if err := s.postRepo.ReorderWhitenestChapters(items); err != nil {
		return fmt.Errorf("failed to reorder chapters: %w", err)
	}
	return nil
}

// validateReorder ensures the supplied list covers numbers 1..N exactly once.
// We require contiguous numbering (rather than allowing gaps) so the resulting
// state matches what the rest of the system already assumes — chapters are a
// dense series, not a sparse one.
func validateReorder(order []dtos.ChapterOrderItem) error {
	if len(order) == 0 {
		return fmt.Errorf("order list must not be empty")
	}
	nums := make(map[int]struct{}, len(order))
	for _, item := range order {
		if _, dup := nums[item.Number]; dup {
			return fmt.Errorf("duplicate chapter number in order: %d", item.Number)
		}
		nums[item.Number] = struct{}{}
	}
	for i := 1; i <= len(order); i++ {
		if _, ok := nums[i]; !ok {
			return fmt.Errorf("chapter numbers must be contiguous 1..%d, missing %d", len(order), i)
		}
	}
	return nil
}
