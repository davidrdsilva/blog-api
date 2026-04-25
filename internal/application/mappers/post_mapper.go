package mappers

import (
	"time"

	"github.com/davidrdsilva/blog-api/internal/application/dtos"
	"github.com/davidrdsilva/blog-api/internal/domain/models"
)

// ToPostResponse converts a domain Post to a PostResponse DTO
func ToPostResponse(post *models.Post) dtos.PostResponse {
	var categoryDTO *dtos.CategoryResponse
	if post.Category != nil {
		c := ToCategoryResponse(post.Category)
		categoryDTO = &c
	}

	tags := ToTagResponses(post.Tags)
	if tags == nil {
		tags = []dtos.TagResponse{}
	}

	return dtos.PostResponse{
		ID:                     post.ID,
		Title:                  post.Title,
		Subtitle:               post.Subtitle,
		Description:            post.Description,
		Image:                  post.Image,
		Date:                   post.Date.In(brt).Format(time.RFC3339),
		Author:                 post.Author,
		Content:                post.Content,
		CategoryID:             post.CategoryID,
		Category:               categoryDTO,
		Tags:                   tags,
		TotalViews:             post.TotalViews,
		WhitenestChapterNumber: post.WhitenestChapterNumber,
		CreatedAt:              post.CreatedAt.In(brt).Format(time.RFC3339),
		UpdatedAt:              post.UpdatedAt.In(brt).Format(time.RFC3339),
	}
}

// Returns nil when post is nil or not a chapter.
func ToWhitenestChapterRef(post *models.Post) *dtos.WhitenestChapterRef {
	if post == nil || post.WhitenestChapterNumber == nil {
		return nil
	}
	return &dtos.WhitenestChapterRef{
		ID:                     post.ID,
		Title:                  post.Title,
		WhitenestChapterNumber: *post.WhitenestChapterNumber,
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
	postDate := time.Now()
	if req.Date != nil {
		postDate = *req.Date
	}

	return &models.Post{
		Title:                  req.Title,
		Subtitle:               req.Subtitle,
		Description:            req.Description,
		Image:                  req.Image,
		Author:                 req.Author,
		Content:                req.Content,
		Date:                   postDate,
		UpdatedAt:              postDate,
		CategoryID:             req.CategoryID,
		WhitenestChapterNumber: req.WhitenestChapterNumber,
	}
}

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
	if req.Date != nil {
		post.Date = *req.Date
	}
	if req.CategoryID != nil {
		post.CategoryID = *req.CategoryID
	}
	if req.WhitenestChapterNumber != nil {
		post.WhitenestChapterNumber = req.WhitenestChapterNumber
	}
}
