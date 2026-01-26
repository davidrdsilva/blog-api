package services

import (
	"io"

	"github.com/davidrdsilva/blog-api/internal/application/dtos"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/storage"
)

// UploadService handles file upload operations
type UploadService struct {
	storage *storage.MinIOStorage
}

// NewUploadService creates a new upload service
func NewUploadService(storage *storage.MinIOStorage) *UploadService {
	return &UploadService{
		storage: storage,
	}
}

// UploadImage handles image upload and returns Editor.js compatible response
func (s *UploadService) UploadImage(file io.Reader, filename string, contentType string) (*dtos.EditorJsUploadResponse, error) {
	// Read file data
	data, err := io.ReadAll(file)
	if err != nil {
		return &dtos.EditorJsUploadResponse{
			Success: 0,
			Error: &dtos.EditorJsErrorDetail{
				Code:    "READ_ERROR",
				Message: "Failed to read file data",
			},
		}, nil
	}

	// Upload to storage
	url, err := s.storage.UploadImage(data, filename, contentType)
	if err != nil {
		// Determine appropriate error code
		errCode := "UPLOAD_FAILED"
		if contains(err.Error(), "file size") {
			errCode = "FILE_TOO_LARGE"
		} else if contains(err.Error(), "invalid file type") {
			errCode = "INVALID_FILE_TYPE"
		} else if contains(err.Error(), "dimensions") {
			errCode = "IMAGE_TOO_LARGE"
		}

		return &dtos.EditorJsUploadResponse{
			Success: 0,
			Error: &dtos.EditorJsErrorDetail{
				Code:    errCode,
				Message: err.Error(),
			},
		}, nil
	}

	// Return success response
	return &dtos.EditorJsUploadResponse{
		Success: 1,
		File: &dtos.EditorJsFileInfo{
			URL: url,
		},
	}, nil
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && hasSubstring(s, substr)))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
