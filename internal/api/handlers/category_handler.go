package handlers

import (
	"net/http"

	"github.com/davidrdsilva/blog-api/internal/application/dtos"
	"github.com/davidrdsilva/blog-api/internal/application/services"
	"github.com/davidrdsilva/blog-api/internal/domain/models"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
	"github.com/gin-gonic/gin"
)

// CategoryHandler handles category-related HTTP requests
type CategoryHandler struct {
	service *services.CategoryService
	logger  *logging.Logger
}

// NewCategoryHandler creates a new category handler
func NewCategoryHandler(service *services.CategoryService, logger *logging.Logger) *CategoryHandler {
	return &CategoryHandler{service: service, logger: logger}
}

// ListCategories handles GET /api/categories
//
// @Summary      List categories
// @Tags         categories
// @Produce      json
// @Param        search  query     string  false  "Case-insensitive name search"
// @Success      200     {object}  dtos.CategoryListResponse
// @Failure      500     {object}  dtos.ErrorResponse
// @Router       /categories [get]
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	filters := models.CategoryFilters{
		Search: c.Query("search"),
	}

	resp, err := h.service.ListCategories(filters)
	if err != nil {
		h.logger.Error("Failed to list categories", logging.F("error", err.Error()))
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to fetch categories",
			},
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// CountPostsByCategory handles GET /api/posts/count/by-category
//
// @Summary      Count posts grouped by category
// @Tags         categories
// @Produce      json
// @Success      200  {object}  dtos.CategoryCountListResponse
// @Failure      500  {object}  dtos.ErrorResponse
// @Router       /posts/count/by-category [get]
func (h *CategoryHandler) CountPostsByCategory(c *gin.Context) {
	resp, err := h.service.CountPostsByCategory()
	if err != nil {
		h.logger.Error("Failed to count posts by category", logging.F("error", err.Error()))
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to count posts by category",
			},
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
