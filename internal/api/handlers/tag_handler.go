package handlers

import (
	"net/http"

	"github.com/davidrdsilva/blog-api/internal/application/dtos"
	"github.com/davidrdsilva/blog-api/internal/application/services"
	"github.com/davidrdsilva/blog-api/internal/domain/models"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
	"github.com/gin-gonic/gin"
)

// TagHandler handles tag-related HTTP requests
type TagHandler struct {
	service *services.TagService
	logger  *logging.Logger
}

// NewTagHandler creates a new tag handler
func NewTagHandler(service *services.TagService, logger *logging.Logger) *TagHandler {
	return &TagHandler{service: service, logger: logger}
}

// ListTags handles GET /api/tags
//
// @Summary      List tags
// @Tags         tags
// @Produce      json
// @Param        search  query     string  false  "Case-insensitive name search"
// @Success      200     {object}  dtos.TagListResponse
// @Failure      500     {object}  dtos.ErrorResponse
// @Router       /tags [get]
func (h *TagHandler) ListTags(c *gin.Context) {
	filters := models.TagFilters{
		Search: c.Query("search"),
	}

	resp, err := h.service.ListTags(filters)
	if err != nil {
		h.logger.Error("Failed to list tags", logging.F("error", err.Error()))
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to fetch tags",
			},
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
