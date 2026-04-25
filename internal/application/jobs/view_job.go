package jobs

// IncrementPostViewsJob carries the post ID whose total_views should be
// incremented by 1. Pushed onto a buffered channel by PostService.GetPost so
// the read response isn't slowed down (or failed) by a write to total_views.
type IncrementPostViewsJob struct {
	PostID string
}
