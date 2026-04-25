// @title           Blog API
// @version         1.0
// @description     RESTful backend for a blog client. Handles posts, image uploads, and URL metadata extraction for Editor.js integration.
// @basePath        /api

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/davidrdsilva/blog-api/docs"
	"github.com/davidrdsilva/blog-api/config"
	"github.com/davidrdsilva/blog-api/internal/api/handlers"
	"github.com/davidrdsilva/blog-api/internal/api/router"
	"github.com/davidrdsilva/blog-api/internal/application/jobs"
	"github.com/davidrdsilva/blog-api/internal/application/services"
	"github.com/davidrdsilva/blog-api/internal/application/workers"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/ai"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/database"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/repository"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/storage"
)

func main() {
	// Root context cancelled on shutdown signal — propagates to background workers.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
	commentRepo := repository.NewPostgresCommentRepository(db)
	categoryRepo := repository.NewPostgresCategoryRepository(db)
	tagRepo := repository.NewPostgresTagRepository(db)

	// Set up the AI comment generation pipeline:
	//   PostService -> jobCh -> CommentWorker -> AICommentService -> Gemini (Ollama fallback) -> DB
	jobCh := make(chan jobs.GenerateCommentsJob, 100)
	ollamaClient := ai.NewOllamaClient(cfg, logger)

	var aiClient ai.AIClient = ollamaClient
	if cfg.Gemini.APIKey != "" {
		geminiClient, err := ai.NewGeminiClient(cfg, logger)
		if err != nil {
			logger.Warn("Failed to initialise Gemini client, falling back to Ollama only",
				logging.F("error", err.Error()),
			)
		} else {
			logger.Info("Gemini client initialised", logging.F("model", cfg.Gemini.Model))
			aiClient = ai.NewFallbackClient(geminiClient, ollamaClient, logger)
		}
	} else {
		logger.Info("GEMINI_API_KEY not set, using Ollama only")
	}

	aiCommentService := services.NewAICommentService(aiClient, commentRepo, logger)
	commentWorker := workers.NewCommentWorker(jobCh, aiCommentService, logger)
	commentWorker.Start(ctx)

	// Initialize services
	postService := services.NewPostService(postRepo, categoryRepo, tagRepo, cfg, jobCh, logger)
	uploadService := services.NewUploadService(minioStorage)
	urlService := services.NewURLService()
	commentService := services.NewCommentService(commentRepo, cfg)
	categoryService := services.NewCategoryService(categoryRepo)
	tagService := services.NewTagService(tagRepo)

	// Initialize handlers
	postHandler := handlers.NewPostHandler(postService, logger)
	uploadHandler := handlers.NewUploadHandler(uploadService, logger)
	urlHandler := handlers.NewURLHandler(urlService, logger)
	commentHandler := handlers.NewCommentHandler(commentService, logger)
	categoryHandler := handlers.NewCategoryHandler(categoryService, logger)
	tagHandler := handlers.NewTagHandler(tagService, logger)

	// Setup router
	r := router.SetupRouter(
		postHandler,
		uploadHandler,
		urlHandler,
		commentHandler,
		categoryHandler,
		tagHandler,
		logger,
		cfg.Server.CORSOrigins,
	)

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

	// Stop accepting new HTTP requests
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", logging.F("error", err.Error()))
		os.Exit(1)
	}

	// Signal the comment worker to stop and close the job channel
	cancel()
	close(jobCh)

	logger.Info("Server exited gracefully")
}
