package services

import (
	"fmt"

	"github.com/davidrdsilva/blog-api/internal/application/dtos"
	"github.com/davidrdsilva/blog-api/internal/application/jobs"
	"github.com/davidrdsilva/blog-api/internal/application/mappers"
	"github.com/davidrdsilva/blog-api/internal/domain/repositories"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
)

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
	}, nil
}
