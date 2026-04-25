package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/davidrdsilva/blog-api/internal/application/dtos"
	"github.com/davidrdsilva/blog-api/internal/application/services"
	"github.com/davidrdsilva/blog-api/internal/domain/models"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PostHandler handles post-related HTTP requests
type PostHandler struct {
	service *services.PostService
	logger  *logging.Logger
}

// NewPostHandler creates a new post handler
func NewPostHandler(service *services.PostService, logger *logging.Logger) *PostHandler {
	return &PostHandler{
		service: service,
		logger:  logger,
	}
}

// CreatePost handles POST /api/posts
//
// @Summary      Create a post
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        post  body      dtos.CreatePostRequest  true  "Post payload"
// @Success      201   {object}  dtos.SuccessResponse
// @Failure      400   {object}  dtos.ErrorResponse
// @Failure      500   {object}  dtos.ErrorResponse
// @Router       /posts [post]
func (h *PostHandler) CreatePost(c *gin.Context) {
	var req dtos.CreatePostRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: "Request validation failed",
				Details: parseValidationErrors(err),
			},
		})
		return
	}

	post, err := h.service.CreatePost(req)
	if err != nil {
		h.logger.Error("Failed to create post", logging.F("error", err.Error()))

		// Check for specific error types
		if containsStr(err.Error(), "invalid image URL") {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Error: dtos.ErrorDetail{
					Code:    "INVALID_IMAGE_URL",
					Message: err.Error(),
				},
			})
			return
		}

		if containsStr(err.Error(), "invalid category") {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Error: dtos.ErrorDetail{
					Code:    "INVALID_CATEGORY",
					Message: err.Error(),
				},
			})
			return
		}

		if containsStr(err.Error(), "whitenest invariant") {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Error: dtos.ErrorDetail{
					Code:    "WHITENEST_INVARIANT_VIOLATION",
					Message: err.Error(),
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to create post",
			},
		})
		return
	}

	h.logger.Info("Post created successfully", logging.F("id", post.ID))
	c.JSON(http.StatusCreated, dtos.SuccessResponse{Data: post})
}

// GetPost handles GET /api/posts/:id
//
// @Summary      Get a post by ID
// @Tags         posts
// @Produce      json
// @Param        id   path      string  true  "Post UUID"
// @Success      200  {object}  dtos.SuccessResponse
// @Failure      400  {object}  dtos.ErrorResponse
// @Failure      404  {object}  dtos.ErrorResponse
// @Failure      500  {object}  dtos.ErrorResponse
// @Router       /posts/{id} [get]
func (h *PostHandler) GetPost(c *gin.Context) {
	id := c.Param("id")

	post, err := h.service.GetPost(id)
	if err != nil {
		if containsStr(err.Error(), "invalid input syntax for type uuid") {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Error: dtos.ErrorDetail{
					Code:    "INVALID_POST_ID",
					Message: "Invalid UUID format",
				},
			})
			return
		}

		h.logger.Error("Failed to fetch post", logging.F("error", err.Error()), logging.F("id", id))
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to fetch post",
			},
		})
		return
	}

	if post == nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{
				Code:    "POST_NOT_FOUND",
				Message: "Post with specified ID does not exist",
			},
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{Data: post})
}

// ListPosts handles GET /api/posts
//
// @Summary      List posts
// @Tags         posts
// @Produce      json
// @Param        search       query     string    false  "Search term"
// @Param        author       query     string    false  "Filter by author"
// @Param        category_id  query     int       false  "Filter by category ID"
// @Param        tags         query     []string  false  "Filter by tag names (OR semantics; repeat the param or comma-join)"  collectionFormat(multi)
// @Param        sortBy       query     string    false  "Sort field"  Enums(date, title, createdAt, updatedAt)
// @Param        sortOrder    query     string    false  "asc or desc"  Enums(asc, desc)
// @Param        page         query     int       false  "Page number (default 1)"
// @Param        limit        query     int       false  "Items per page (default 6, max 50)"
// @Success      200          {object}  dtos.PostListResponse
// @Failure      500          {object}  dtos.ErrorResponse
// @Router       /posts [get]
func (h *PostHandler) ListPosts(c *gin.Context) {
	// Parse query parameters
	filters := models.PostFilters{
		Search:    c.Query("search"),
		Author:    c.Query("author"),
		SortBy:    c.Query("sortBy"),
		SortOrder: c.Query("sortOrder"),
		Page:      parseIntQuery(c, "page", 1),
		Limit:     parseIntQuery(c, "limit", 6),
		TagNames:  parseTagsQuery(c),
	}

	if categoryStr := c.Query("category_id"); categoryStr != "" {
		if cid, err := strconv.Atoi(categoryStr); err == nil && cid > 0 {
			filters.CategoryID = &cid
		}
	}

	// Whitenest chapters live behind /api/whitenest/chapters/:number — exclude
	// them from the generic feed unless the caller explicitly opts in.
	excludeChapters := false
	filters.IsWhitenestChapter = &excludeChapters
	if v := c.Query("is_whitenest_chapter"); v != "" {
		if parsed, err := strconv.ParseBool(v); err == nil {
			filters.IsWhitenestChapter = &parsed
		}
	}

	posts, err := h.service.ListPosts(filters)
	if err != nil {
		h.logger.Error("Failed to list posts", logging.F("error", err.Error()))
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to fetch posts",
			},
		})
		return
	}

	c.JSON(http.StatusOK, posts)
}

// Similar handles GET /api/posts/:id/similar
//
// @Summary      List posts similar to the given post
// @Tags         posts
// @Produce      json
// @Param        id   path      string  true  "Post UUID"
// @Success      200  {object}  dtos.PostListResponse
// @Failure      400  {object}  dtos.ErrorResponse
// @Failure      500  {object}  dtos.ErrorResponse
// @Router       /posts/{id}/similar [get]
func (h *PostHandler) Similar(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.service.GetSimilarPosts(id, 5)
	if err != nil {
		if containsStr(err.Error(), "invalid UUID") {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Error: dtos.ErrorDetail{
					Code:    "INVALID_POST_ID",
					Message: "Invalid UUID format",
				},
			})
			return
		}
		h.logger.Error("Failed to fetch similar posts",
			logging.F("error", err.Error()),
			logging.F("id", id),
		)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to fetch similar posts",
			},
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// MostViewed handles GET /api/posts/most-viewed
//
// @Summary      List the most viewed posts
// @Tags         posts
// @Produce      json
// @Success      200  {object}  dtos.PostListResponse
// @Failure      500  {object}  dtos.ErrorResponse
// @Router       /posts/most-viewed [get]
func (h *PostHandler) MostViewed(c *gin.Context) {
	resp, err := h.service.GetMostViewed(5)
	if err != nil {
		h.logger.Error("Failed to fetch most viewed posts", logging.F("error", err.Error()))
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to fetch most viewed posts",
			},
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UpdatePost handles PUT /api/posts/:id
//
// @Summary      Update a post
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        id    path      string                  true  "Post UUID"
// @Param        post  body      dtos.UpdatePostRequest  true  "Fields to update"
// @Success      200   {object}  dtos.SuccessResponse
// @Failure      400   {object}  dtos.ErrorResponse
// @Failure      404   {object}  dtos.ErrorResponse
// @Failure      500   {object}  dtos.ErrorResponse
// @Router       /posts/{id} [put]
func (h *PostHandler) UpdatePost(c *gin.Context) {
	id := c.Param("id")
	var req dtos.UpdatePostRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: "Request validation failed",
				Details: parseValidationErrors(err),
			},
		})
		return
	}

	post, err := h.service.UpdatePost(id, req)
	if err != nil {
		if containsStr(err.Error(), "invalid UUID") {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Error: dtos.ErrorDetail{
					Code:    "INVALID_POST_ID",
					Message: "Invalid UUID format",
				},
			})
			return
		}

		if containsStr(err.Error(), "invalid image URL") {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Error: dtos.ErrorDetail{
					Code:    "INVALID_IMAGE_URL",
					Message: err.Error(),
				},
			})
			return
		}

		if containsStr(err.Error(), "invalid category") {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Error: dtos.ErrorDetail{
					Code:    "INVALID_CATEGORY",
					Message: err.Error(),
				},
			})
			return
		}

		if containsStr(err.Error(), "whitenest invariant") {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Error: dtos.ErrorDetail{
					Code:    "WHITENEST_INVARIANT_VIOLATION",
					Message: err.Error(),
				},
			})
			return
		}

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{
				Error: dtos.ErrorDetail{
					Code:    "POST_NOT_FOUND",
					Message: "Post with specified ID does not exist",
				},
			})
			return
		}

		h.logger.Error("Failed to update post", logging.F("error", err.Error()), logging.F("id", id))
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to update post",
			},
		})
		return
	}

	if post == nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{
				Code:    "POST_NOT_FOUND",
				Message: "Post with specified ID does not exist",
			},
		})
		return
	}

	h.logger.Info("Post updated successfully", logging.F("id", id))
	c.JSON(http.StatusOK, dtos.SuccessResponse{Data: post})
}

// DeletePost handles DELETE /api/posts/:id
//
// @Summary      Delete a post
// @Tags         posts
// @Param        id  path  string  true  "Post UUID"
// @Success      204
// @Failure      400  {object}  dtos.ErrorResponse
// @Failure      404  {object}  dtos.ErrorResponse
// @Failure      500  {object}  dtos.ErrorResponse
// @Router       /posts/{id} [delete]
func (h *PostHandler) DeletePost(c *gin.Context) {
	id := c.Param("id")

	err := h.service.DeletePost(id)
	if err != nil {
		if containsStr(err.Error(), "invalid UUID") {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Error: dtos.ErrorDetail{
					Code:    "INVALID_POST_ID",
					Message: "Invalid UUID format",
				},
			})
			return
		}

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{
				Error: dtos.ErrorDetail{
					Code:    "POST_NOT_FOUND",
					Message: "Post with specified ID does not exist",
				},
			})
			return
		}

		h.logger.Error("Failed to delete post", logging.F("error", err.Error()), logging.F("id", id))
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to delete post",
			},
		})
		return
	}

	h.logger.Info("Post deleted successfully", logging.F("id", id))
	c.Status(http.StatusNoContent)
}

// parseTagsQuery returns the tag-name filter values. Accepts repeated
// `?tags=foo&tags=bar` plus a comma-joined `?tags=foo,bar` form for convenience.
func parseTagsQuery(c *gin.Context) []string {
	raw := c.QueryArray("tags")
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		for _, p := range strings.Split(v, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
	}
	return out
}

// parseIntQuery parses an integer query parameter with a default value
func parseIntQuery(c *gin.Context, key string, defaultValue int) int {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// parseValidationErrors parses validation errors into a map
func parseValidationErrors(err error) map[string][]string {
	// Simple error message for now
	// Could be enhanced with gin validator for field-specific errors
	return map[string][]string{
		"_error": {err.Error()},
	}
}

// containsStr checks if a string contains a substring
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && hasSubstr(s, substr)
}

func hasSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
