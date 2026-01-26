package handlers

import (
	"net/http"

	"github.com/davidrdsilva/blog-api/internal/application/services"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
	"github.com/gin-gonic/gin"
)

// UploadHandler handles file upload requests
type UploadHandler struct {
	service *services.UploadService
	logger  *logging.Logger
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(service *services.UploadService, logger *logging.Logger) *UploadHandler {
	return &UploadHandler{
		service: service,
		logger:  logger,
	}
}

// UploadImage handles POST /api/upload
func (h *UploadHandler) UploadImage(c *gin.Context) {
	// Get file from multipart form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		h.logger.Warn("No file provided in upload request")
		c.JSON(http.StatusOK, map[string]interface{}{
			"success": 0,
			"error": map[string]string{
				"code":    "NO_FILE_PROVIDED",
				"message": "No file in request",
			},
		})
		return
	}
	defer file.Close()

	// Get content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	h.logger.Info("Processing file upload",
		logging.F("filename", header.Filename),
		logging.F("content_type", contentType),
		logging.F("size_bytes", header.Size),
	)

	// Upload file
	response, err := h.service.UploadImage(file, header.Filename, contentType)
	if err != nil {
		h.logger.Error("Failed to upload image", logging.F("error", err.Error()))
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": 0,
			"error": map[string]string{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to process upload",
			},
		})
		return
	}

	// Log success or failure
	if response.Success == 1 {
		h.logger.Info("Image uploaded successfully", logging.F("url", response.File.URL))
	} else {
		h.logger.Warn("Upload validation failed", logging.F("error", response.Error.Code))
	}

	c.JSON(http.StatusOK, response)
}
