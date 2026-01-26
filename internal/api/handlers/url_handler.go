package handlers

import (
	"net/http"

	"github.com/davidrdsilva/blog-api/internal/application/services"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
	"github.com/gin-gonic/gin"
)

// URLHandler handles URL metadata fetching requests
type URLHandler struct {
	service *services.URLService
	logger  *logging.Logger
}

// NewURLHandler creates a new URL handler
func NewURLHandler(service *services.URLService, logger *logging.Logger) *URLHandler {
	return &URLHandler{
		service: service,
		logger:  logger,
	}
}

// FetchURLMetadata handles GET /api/fetch-url
func (h *URLHandler) FetchURLMetadata(c *gin.Context) {
	url := c.Query("url")

	if url == "" {
		c.JSON(http.StatusOK, map[string]interface{}{
			"success": 0,
			"error": map[string]string{
				"code":    "INVALID_URL",
				"message": "URL parameter is missing or malformed",
			},
		})
		return
	}

	h.logger.Info("Fetching URL metadata", logging.F("url", url))

	response, err := h.service.FetchURLMetadata(url)
	if err != nil {
		h.logger.Error("Failed to fetch URL metadata", logging.F("error", err.Error()))
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": 0,
			"error": map[string]string{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch URL metadata",
			},
		})
		return
	}

	if response.Success == 1 {
		h.logger.Info("URL metadata fetched successfully", logging.F("title", response.Meta.Title))
	} else {
		h.logger.Warn("URL metadata fetch failed", logging.F("error", response.Error.Code))
	}

	c.JSON(http.StatusOK, response)
}
