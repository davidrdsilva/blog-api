package database

import (
	"fmt"

	"github.com/davidrdsilva/blog-api/internal/domain/models"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(dsn string, log *logging.Logger) (*gorm.DB, error) {
	log.Info("Connecting to PostgreSQL database...")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Use custom logger instead
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Info("Successfully connected to PostgreSQL")

	// Get underlying SQL DB for connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	log.Info("Database connection pool configured")

	return db, nil
}

// RunMigrations executes database migrations
func RunMigrations(db *gorm.DB, log *logging.Logger) error {
	log.Info("Running database migrations...")

	err := db.AutoMigrate(
		&models.Post{},
		&models.Comment{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Info("Database migrations completed successfully")

	// Create indexes for better query performance
	if err := createIndexes(db, log); err != nil {
		return err
	}

	return nil
}

// createIndexes creates database indexes for optimized queries
func createIndexes(db *gorm.DB, log *logging.Logger) error {
	log.Info("Creating database indexes...")

	// Index on date for sorting
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_posts_date ON posts(date DESC)").Error; err != nil {
		return fmt.Errorf("failed to create date index: %w", err)
	}

	// Index on author for filtering
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_posts_author ON posts(author)").Error; err != nil {
		return fmt.Errorf("failed to create author index: %w", err)
	}

	// Index on post_id for comments
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id)").Error; err != nil {
		return fmt.Errorf("failed to create post_id index: %w", err)
	}

	// Full-text search index on searchable fields
	// Create a computed column for full-text search
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_posts_search ON posts 
		USING GIN(to_tsvector('english', 
			title || ' ' || COALESCE(subtitle, '') || ' ' || description
		))
	`).Error; err != nil {
		return fmt.Errorf("failed to create search index: %w", err)
	}

	log.Info("Database indexes created successfully")
	return nil
}
