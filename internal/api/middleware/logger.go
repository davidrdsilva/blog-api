package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/davidrdsilva/blog-api/internal/application/dtos"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
	"github.com/gin-gonic/gin"
)

// Logger middleware logs HTTP requests with colored output
func Logger(log *logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// Log with appropriate level based on status code
		fields := []logging.Field{
			logging.F("method", method),
			logging.F("path", path),
			logging.F("status", statusCode),
			logging.F("duration", fmt.Sprintf("%v", duration)),
		}

		message := fmt.Sprintf("[%s %s]", method, path)

		if statusCode >= 500 {
			log.Error(message, fields...)
		} else if statusCode >= 400 {
			log.Warn(message, fields...)
		} else {
			log.Info(message, fields...)
		}
	}
}

// ErrorHandler middleware handles panics and returns standardized error responses
func ErrorHandler(log *logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error("Panic recovered",
					logging.F("error", fmt.Sprintf("%v", err)),
					logging.F("path", c.Request.URL.Path),
				)

				c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
					Error: dtos.ErrorDetail{
						Code:    "INTERNAL_ERROR",
						Message: "An internal server error occurred",
					},
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}
