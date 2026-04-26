package router

import (
	"github.com/davidrdsilva/blog-api/internal/api/handlers"
	"github.com/davidrdsilva/blog-api/internal/api/middleware"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter configures all routes and middleware
func SetupRouter(
	postHandler *handlers.PostHandler,
	uploadHandler *handlers.UploadHandler,
	urlHandler *handlers.URLHandler,
	commentHandler *handlers.CommentHandler,
	categoryHandler *handlers.CategoryHandler,
	tagHandler *handlers.TagHandler,
	whitenestHandler *handlers.WhitenestHandler,
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
		// Static segments must precede the :id route to avoid being shadowed.
		api.GET("/posts/count/by-category", categoryHandler.CountPostsByCategory)
		api.GET("/posts/most-viewed", postHandler.MostViewed)
		api.GET("/posts/:id", postHandler.GetPost)
		api.GET("/posts/:id/similar", postHandler.Similar)
		api.POST("/posts", postHandler.CreatePost)
		api.PUT("/posts/:id", postHandler.UpdatePost)
		api.DELETE("/posts/:id", postHandler.DeletePost)

		// Comment endpoints
		api.POST("/comments", commentHandler.CreateComment)
		api.GET("/comments", commentHandler.ListComments)
		api.DELETE("/comments/:id", commentHandler.DeleteComment)

		// Category and tag endpoints
		api.GET("/categories", categoryHandler.ListCategories)
		api.GET("/tags", tagHandler.ListTags)

		// Whitenest serial-fiction endpoints
		api.GET("/whitenest/chapters", whitenestHandler.ListChapters)
		api.GET("/whitenest/chapters/:number", whitenestHandler.GetChapter)

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

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
