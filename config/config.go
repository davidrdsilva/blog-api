package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	Database DatabaseConfig
	MinIO    MinIOConfig
	Server   ServerConfig
	Upload   UploadConfig
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// MinIOConfig holds MinIO object storage settings
type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	PublicURL string
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Port        string
	CORSOrigins []string
}

// UploadConfig holds file upload constraints
type UploadConfig struct {
	MaxFileSizeMB      int
	MaxImageDimension  int
	AllowedMimeTypes   []string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	maxFileSize, err := strconv.Atoi(getEnv("MAX_FILE_SIZE_MB", "5"))
	if err != nil {
		return nil, fmt.Errorf("invalid MAX_FILE_SIZE_MB: %w", err)
	}

	maxImageDim, err := strconv.Atoi(getEnv("MAX_IMAGE_DIMENSION", "4096"))
	if err != nil {
		return nil, fmt.Errorf("invalid MAX_IMAGE_DIMENSION: %w", err)
	}

	useSSL := getEnv("MINIO_USE_SSL", "false") == "true"

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "bloguser"),
			Password: getEnv("DB_PASSWORD", "blogpassword"),
			DBName:   getEnv("DB_NAME", "blogdb"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		MinIO: MinIOConfig{
			Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			Bucket:    getEnv("MINIO_BUCKET", "blog"),
			UseSSL:    useSSL,
			PublicURL: getEnv("MINIO_PUBLIC_URL", "http://localhost:9000"),
		},
		Server: ServerConfig{
			Port:        getEnv("SERVER_PORT", "8080"),
			CORSOrigins: parseCommaSeparated(getEnv("CORS_ORIGINS", "http://localhost:3000")),
		},
		Upload: UploadConfig{
			MaxFileSizeMB:     maxFileSize,
			MaxImageDimension: maxImageDim,
			AllowedMimeTypes:  []string{"image/jpeg", "image/png", "image/gif", "image/webp"},
		},
	}, nil
}

// GetDSN returns the PostgreSQL connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseCommaSeparated splits a comma-separated string into a slice
func parseCommaSeparated(s string) []string {
	if s == "" {
		return []string{}
	}
	
	var result []string
	current := ""
	
	for _, char := range s {
		if char == ',' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	
	if current != "" {
		result = append(result, current)
	}
	
	return result
}
