package handlers

import (
	"net/http"
	"strconv"

	"github.com/davidrdsilva/blog-api/internal/application/dtos"
	"github.com/davidrdsilva/blog-api/internal/application/services"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
	"github.com/gin-gonic/gin"
)

type WhitenestHandler struct {
	service *services.WhitenestService
	logger  *logging.Logger
}

func NewWhitenestHandler(service *services.WhitenestService, logger *logging.Logger) *WhitenestHandler {
	return &WhitenestHandler{
		service: service,
		logger:  logger,
	}
}

// GetChapter handles GET /api/whitenest/chapters/:number
//
// @Summary      Get a Whitenest chapter by number
// @Description  Returns the chapter with the given serial number along with
// @Description  minimal references to the previous and next chapters, if any.
// @Tags         whitenest
// @Produce      json
// @Param        number  path      int  true  "Chapter number (1-indexed)"
// @Success      200     {object}  dtos.SuccessResponse{data=dtos.WhitenestChapterResponse}
// @Failure      400     {object}  dtos.ErrorResponse
// @Failure      404     {object}  dtos.ErrorResponse
// @Failure      500     {object}  dtos.ErrorResponse
// @Router       /whitenest/chapters/{number} [get]
func (h *WhitenestHandler) GetChapter(c *gin.Context) {
	raw := c.Param("number")
	number, err := strconv.Atoi(raw)
	if err != nil || number < 1 {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{
				Code:    "INVALID_CHAPTER_NUMBER",
				Message: "Chapter number must be a positive integer",
			},
		})
		return
	}

	resp, err := h.service.GetChapterByNumber(number)
	if err != nil {
		h.logger.Error("Failed to fetch Whitenest chapter",
			logging.F("error", err.Error()),
			logging.F("number", number),
		)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to fetch chapter",
			},
		})
		return
	}

	if resp == nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{
				Code:    "CHAPTER_NOT_FOUND",
				Message: "No Whitenest chapter exists with that number",
			},
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{Data: resp})
}

// ReorderChapters handles PUT /api/whitenest/chapters/order
//
// @Summary      Reorder Whitenest chapters
// @Description  Accepts the full ordered list of (post_id, number) pairs and
// @Description  rewrites chapter numbers atomically. The submitted set must
// @Description  cover every existing chapter exactly once with contiguous
// @Description  numbers 1..N. A mismatch (e.g. concurrent publish/unpublish)
// @Description  returns 409 so the client can refresh and retry.
// @Tags         whitenest
// @Accept       json
// @Produce      json
// @Param        body  body      dtos.ReorderChaptersRequest  true  "Full chapter order"
// @Success      204
// @Failure      400  {object}  dtos.ErrorResponse
// @Failure      409  {object}  dtos.ErrorResponse
// @Failure      500  {object}  dtos.ErrorResponse
// @Router       /whitenest/chapters/order [put]
func (h *WhitenestHandler) ReorderChapters(c *gin.Context) {
	var req dtos.ReorderChaptersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	if err := h.service.ReorderChapters(req); err != nil {
		msg := err.Error()
		switch {
		case containsStr(msg, "chapter set mismatch"):
			c.JSON(http.StatusConflict, dtos.ErrorResponse{
				Error: dtos.ErrorDetail{
					Code:    "CHAPTER_SET_MISMATCH",
					Message: msg,
				},
			})
			return
		case containsStr(msg, "duplicate") || containsStr(msg, "must be contiguous") || containsStr(msg, "must not be empty"):
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Error: dtos.ErrorDetail{
					Code:    "INVALID_CHAPTER_ORDER",
					Message: msg,
				},
			})
			return
		}
		h.logger.Error("Failed to reorder Whitenest chapters",
			logging.F("error", msg),
		)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to reorder chapters",
			},
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListChapters handles GET /api/whitenest/chapters
//
// @Summary      List all Whitenest chapters
// @Description  Returns every Whitenest chapter ordered by chapter number ASC
// @Description  with the lightweight fields needed for list views (id, title,
// @Description  image, tags, chapter number).
// @Tags         whitenest
// @Produce      json
// @Success      200  {object}  dtos.SuccessResponse{data=[]dtos.WhitenestChapterSummary}
// @Failure      500  {object}  dtos.ErrorResponse
// @Router       /whitenest/chapters [get]
func (h *WhitenestHandler) ListChapters(c *gin.Context) {
	chapters, err := h.service.ListChapters()
	if err != nil {
		h.logger.Error("Failed to list Whitenest chapters",
			logging.F("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to list chapters",
			},
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{Data: chapters})
}
