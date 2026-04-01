package workers

import (
	"context"
	"time"

	"github.com/davidrdsilva/blog-api/internal/application/jobs"
	"github.com/davidrdsilva/blog-api/internal/application/services"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
)

const jobTimeout = 3 * time.Minute

// CommentWorker reads GenerateCommentsJobs from a channel and processes them one at a time.
type CommentWorker struct {
	jobs      <-chan jobs.GenerateCommentsJob
	aiService *services.AICommentService
	logger    *logging.Logger
}

func NewCommentWorker(
	jobCh <-chan jobs.GenerateCommentsJob,
	aiService *services.AICommentService,
	logger *logging.Logger,
) *CommentWorker {
	return &CommentWorker{
		jobs:      jobCh,
		aiService: aiService,
		logger:    logger,
	}
}

// Start launches the worker goroutine. It drains remaining jobs until the channel
// is closed or the parent context is cancelled.
func (w *CommentWorker) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case job, ok := <-w.jobs:
				if !ok {
					w.logger.Info("Comment worker: channel closed, exiting")
					return
				}
				jobCtx, cancel := context.WithTimeout(ctx, jobTimeout)
				if err := w.aiService.GenerateAndSave(jobCtx, job); err != nil {
					w.logger.Error("AI comment generation failed",
						logging.F("postId", job.PostID),
						logging.F("error", err.Error()),
					)
				}
				cancel()
			case <-ctx.Done():
				w.logger.Info("Comment worker: context cancelled, exiting")
				return
			}
		}
	}()
}
