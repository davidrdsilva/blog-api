package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/davidrdsilva/blog-api/config"
	"github.com/davidrdsilva/blog-api/internal/api/handlers"
	"github.com/davidrdsilva/blog-api/internal/api/router"
	"github.com/davidrdsilva/blog-api/internal/application/services"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/database"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/repository"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/storage"
)

func main() {
	// Initialize logger
	logger := logging.NewLogger("blog-api")
	logger.Info("Starting Blog API server...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration", logging.F("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("Configuration loaded successfully")

	// Connect to database
	db, err := database.NewPostgresDB(cfg.GetDSN(), logger)
	if err != nil {
		logger.Error("Failed to connect to database", logging.F("error", err.Error()))
		os.Exit(1)
	}

	// Run migrations
	if err := database.RunMigrations(db, logger); err != nil {
		logger.Error("Failed to run migrations", logging.F("error", err.Error()))
		os.Exit(1)
	}

	// Initialize MinIO storage
	minioStorage, err := storage.NewMinIOStorage(cfg, logger)
	if err != nil {
		logger.Error("Failed to initialize MinIO storage", logging.F("error", err.Error()))
		os.Exit(1)
	}

	// Initialize repositories
	postRepo := repository.NewPostgresPostRepository(db)

	// Initialize services
	postService := services.NewPostService(postRepo, cfg)
	uploadService := services.NewUploadService(minioStorage)
	urlService := services.NewURLService()

	// Initialize handlers
	postHandler := handlers.NewPostHandler(postService, logger)
	uploadHandler := handlers.NewUploadHandler(uploadService, logger)
	urlHandler := handlers.NewURLHandler(urlService, logger)

	// Setup router
	r := router.SetupRouter(postHandler, uploadHandler, urlHandler, logger, cfg.Server.CORSOrigins)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Server listening", logging.F("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start server", logging.F("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", logging.F("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("Server exited gracefully")
}
