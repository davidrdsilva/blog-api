package handlers

import (
	"net/http"

	"github.com/davidrdsilva/blog-api/internal/application/dtos"
	"github.com/davidrdsilva/blog-api/internal/application/services"
	"github.com/davidrdsilva/blog-api/internal/domain/models"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CommentHandler struct {
	service *services.CommentService
	logger  *logging.Logger
}

func NewCommentHandler(service *services.CommentService, logger *logging.Logger) *CommentHandler {
	return &CommentHandler{
		service: service,
		logger:  logger,
	}
}

func (h *CommentHandler) CreateComment(c *gin.Context) {
	var req dtos.CreateCommentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: "Request validation failed",
				Details: parseValidationErrors(err),
			},
		})
		return
	}

	comment, err := h.service.CreateComment(req)
	if err != nil {
		if containsStr(err.Error(), "invalid UUID") {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Error: dtos.ErrorDetail{
					Code:    "INVALID_COMMENT_ID",
					Message: "Invalid UUID format",
				},
			})
			return
		}

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{
				Error: dtos.ErrorDetail{
					Code:    "COMMENT_NOT_FOUND",
					Message: "Comment with specified ID does not exist",
				},
			})
			return
		}

		h.logger.Error("Failed to create comment", logging.F("error", err.Error()))
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to create comment",
			},
		})
		return
	}

	h.logger.Info("Comment created successfully", logging.F("id", comment.ID))
	c.JSON(http.StatusCreated, dtos.SuccessResponse{Data: comment})
}

func (h *CommentHandler) ListComments(c *gin.Context) {
	comments, err := h.service.GetComments(models.CommentFilters{
		PostID:    c.Query("post_id"),
		Author:    c.Query("author"),
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to list comments",
			},
		})
		return
	}

	c.JSON(http.StatusOK, comments)
}

func (h *CommentHandler) DeleteComment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{
				Code:    "INVALID_COMMENT_ID",
				Message: "Invalid comment ID",
			},
		})
		return
	}

	if err := h.service.DeleteComment(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{
				Error: dtos.ErrorDetail{
					Code:    "COMMENT_NOT_FOUND",
					Message: "Comment with specified ID does not exist",
				},
			})
			return
		}

		h.logger.Error("Failed to delete comment", logging.F("error", err.Error()))
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to delete comment",
			},
		})
		return
	}

	h.logger.Info("Comment deleted successfully", logging.F("id", id))
	c.JSON(http.StatusOK, dtos.SuccessResponse{Data: id})
}
