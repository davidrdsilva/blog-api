package workers

import (
	"context"
	"time"

	"github.com/davidrdsilva/blog-api/internal/application/jobs"
	"github.com/davidrdsilva/blog-api/internal/domain/repositories"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
)

const viewJobTimeout = 30 * time.Second

// ViewCounterWorker consumes IncrementPostViewsJobs from a channel and
// increments total_views one row at a time. Decoupled from the read path so
// failures here don't surface to the user.
type ViewCounterWorker struct {
	jobs   <-chan jobs.IncrementPostViewsJob
	repo   repositories.PostRepository
	logger *logging.Logger
}

func NewViewCounterWorker(
	jobCh <-chan jobs.IncrementPostViewsJob,
	repo repositories.PostRepository,
	logger *logging.Logger,
) *ViewCounterWorker {
	return &ViewCounterWorker{
		jobs:   jobCh,
		repo:   repo,
		logger: logger,
	}
}

// Start launches the worker goroutine. It drains remaining jobs until the
// channel is closed or the parent context is cancelled.
func (w *ViewCounterWorker) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case job, ok := <-w.jobs:
				if !ok {
					w.logger.Info("View counter worker: channel closed, exiting")
					return
				}
				_, cancel := context.WithTimeout(ctx, viewJobTimeout)
				if err := w.repo.IncrementViews(job.PostID); err != nil {
					w.logger.Error("Failed to increment post views",
						logging.F("postId", job.PostID),
						logging.F("error", err.Error()),
					)
				}
				cancel()
			case <-ctx.Done():
				w.logger.Info("View counter worker: context cancelled, exiting")
				return
			}
		}
	}()
}
