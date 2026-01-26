package storage

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"path/filepath"
	"strings"
	"time"

	"github.com/davidrdsilva/blog-api/config"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOStorage handles file uploads to MinIO object storage
type MinIOStorage struct {
	client    *minio.Client
	bucket    string
	publicURL string
	config    *config.UploadConfig
	logger    *logging.Logger
}

// NewMinIOStorage creates a new MinIO storage client
func NewMinIOStorage(cfg *config.Config, log *logging.Logger) (*MinIOStorage, error) {
	log.Info("Initializing MinIO client...",
		logging.F("endpoint", cfg.MinIO.Endpoint),
		logging.F("bucket", cfg.MinIO.Bucket),
	)

	client, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, ""),
		Secure: cfg.MinIO.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	storage := &MinIOStorage{
		client:    client,
		bucket:    cfg.MinIO.Bucket,
		publicURL: cfg.MinIO.PublicURL,
		config:    &cfg.Upload,
		logger:    log,
	}

	// Ensure bucket exists
	if err := storage.ensureBucket(); err != nil {
		return nil, err
	}

	log.Info("MinIO client initialized successfully")
	return storage, nil
}

// ensureBucket creates the bucket if it doesn't exist
func (s *MinIOStorage) ensureBucket() error {
	ctx := context.Background()

	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		s.logger.Info("Creating MinIO bucket", logging.F("bucket", s.bucket))

		err = s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}

		// Set bucket policy to public-read for uploaded images
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}]
		}`, s.bucket)

		err = s.client.SetBucketPolicy(ctx, s.bucket, policy)
		if err != nil {
			return fmt.Errorf("failed to set bucket policy: %w", err)
		}

		s.logger.Info("Bucket created with public-read policy")
	}

	return nil
}

// UploadImage uploads an image file and returns its public URL
func (s *MinIOStorage) UploadImage(fileData []byte, originalFilename string, contentType string) (string, error) {
	// Validate MIME type
	if !s.isAllowedMimeType(contentType) {
		return "", fmt.Errorf("invalid file type: %s (allowed: %v)", contentType, s.config.AllowedMimeTypes)
	}

	// Validate file size
	fileSizeMB := float64(len(fileData)) / (1024 * 1024)
	if fileSizeMB > float64(s.config.MaxFileSizeMB) {
		return "", fmt.Errorf("file size %.2fMB exceeds maximum allowed size of %dMB", fileSizeMB, s.config.MaxFileSizeMB)
	}

	// Validate image dimensions
	if err := s.validateImageDimensions(fileData); err != nil {
		return "", err
	}

	// Generate unique filename with UUID to avoid collisions and spaces
	ext := filepath.Ext(originalFilename)
	if ext == "" {
		ext = s.getExtensionFromMimeType(contentType)
	}
	filename := fmt.Sprintf("uploads/%s%s", uuid.New().String(), ext)

	// Upload to MinIO
	ctx := context.Background()
	reader := bytes.NewReader(fileData)

	_, err := s.client.PutObject(ctx, s.bucket, filename, reader, int64(len(fileData)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Generate public URL
	url := fmt.Sprintf("%s/%s/%s", s.publicURL, s.bucket, filename)

	s.logger.Info("Image uploaded successfully",
		logging.F("filename", filename),
		logging.F("size_mb", fmt.Sprintf("%.2f", fileSizeMB)),
	)

	return url, nil
}

// isAllowedMimeType checks if the MIME type is allowed
func (s *MinIOStorage) isAllowedMimeType(mimeType string) bool {
	for _, allowed := range s.config.AllowedMimeTypes {
		if strings.EqualFold(mimeType, allowed) {
			return true
		}
	}
	return false
}

// validateImageDimensions checks if image dimensions are within limits
func (s *MinIOStorage) validateImageDimensions(fileData []byte) error {
	reader := bytes.NewReader(fileData)
	img, _, err := image.DecodeConfig(reader)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	maxDim := s.config.MaxImageDimension
	if img.Width > maxDim || img.Height > maxDim {
		return fmt.Errorf("image dimensions %dx%d exceed maximum allowed %dx%d",
			img.Width, img.Height, maxDim, maxDim)
	}

	return nil
}

// getExtensionFromMimeType returns file extension for a given MIME type
func (s *MinIOStorage) getExtensionFromMimeType(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ".jpg"
	}
}

// HealthCheck verifies MinIO connectivity
func (s *MinIOStorage) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.client.BucketExists(ctx, s.bucket)
	return err
}
