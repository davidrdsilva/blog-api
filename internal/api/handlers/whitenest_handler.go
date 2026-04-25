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
