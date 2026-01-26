package router

import (
	"github.com/davidrdsilva/blog-api/internal/api/handlers"
	"github.com/davidrdsilva/blog-api/internal/api/middleware"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures all routes and middleware
func SetupRouter(
	postHandler *handlers.PostHandler,
	uploadHandler *handlers.UploadHandler,
	urlHandler *handlers.URLHandler,
	logger *logging.Logger,
	corsOrigins []string,
) *gin.Engine {
	// Set Gin to release mode
	gin.SetMode(gin.ReleaseMode)

	// Create router
	r := gin.New()

	// Apply global middleware
	r.Use(middleware.ErrorHandler(logger))
	r.Use(middleware.Logger(logger))
	r.Use(middleware.CORS(corsOrigins))

	// API routes
	api := r.Group("/api")
	{
		// Post endpoints
		api.GET("/posts", postHandler.ListPosts)
		api.GET("/posts/:id", postHandler.GetPost)
		api.POST("/posts", postHandler.CreatePost)
		api.PUT("/posts/:id", postHandler.UpdatePost)
		api.DELETE("/posts/:id", postHandler.DeletePost)

		// Upload endpoint
		api.POST("/upload", uploadHandler.UploadImage)

		// URL metadata endpoint
		api.GET("/fetch-url", urlHandler.FetchURLMetadata)
	}

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	return r
}
