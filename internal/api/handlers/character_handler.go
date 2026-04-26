package handlers

import (
	"errors"
	"net/http"

	"github.com/davidrdsilva/blog-api/internal/application/dtos"
	"github.com/davidrdsilva/blog-api/internal/application/services"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CharacterHandler struct {
	service *services.CharacterService
	logger  *logging.Logger
}

func NewCharacterHandler(service *services.CharacterService, logger *logging.Logger) *CharacterHandler {
	return &CharacterHandler{service: service, logger: logger}
}

// ListCharacters handles GET /api/characters
//
// @Summary      List characters
// @Description  Returns all characters, alphabetized by short name. Optionally
// @Description  filter by `?search=` (case-insensitive on full and short name).
// @Tags         characters
// @Produce      json
// @Param        search  query     string  false  "Substring search on short_name/full_name"
// @Success      200  {object}  dtos.SuccessResponse{data=[]dtos.CharacterResponse}
// @Failure      500  {object}  dtos.ErrorResponse
// @Router       /characters [get]
func (h *CharacterHandler) ListCharacters(c *gin.Context) {
	search := c.Query("search")
	rows, err := h.service.ListCharacters(search)
	if err != nil {
		h.logger.Error("Failed to list characters", logging.F("error", err.Error()))
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{Code: "INTERNAL_ERROR", Message: "Failed to list characters"},
		})
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse{Data: rows})
}

// GetCharacter handles GET /api/characters/:id
//
// @Summary      Get a character by ID
// @Tags         characters
// @Produce      json
// @Param        id   path      string  true  "Character UUID"
// @Success      200  {object}  dtos.SuccessResponse{data=dtos.CharacterResponse}
// @Failure      400  {object}  dtos.ErrorResponse
// @Failure      404  {object}  dtos.ErrorResponse
// @Failure      500  {object}  dtos.ErrorResponse
// @Router       /characters/{id} [get]
func (h *CharacterHandler) GetCharacter(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.service.GetCharacter(id)
	if err != nil {
		if containsStr(err.Error(), "invalid UUID") {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Error: dtos.ErrorDetail{Code: "INVALID_ID", Message: "Invalid character ID"},
			})
			return
		}
		h.logger.Error("Failed to fetch character", logging.F("error", err.Error()))
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{Code: "INTERNAL_ERROR", Message: "Failed to fetch character"},
		})
		return
	}
	if resp == nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{Code: "CHARACTER_NOT_FOUND", Message: "Character not found"},
		})
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse{Data: resp})
}

// CreateCharacter handles POST /api/characters
//
// @Summary      Create a character
// @Tags         characters
// @Accept       json
// @Produce      json
// @Param        character  body      dtos.CreateCharacterRequest  true  "Character payload"
// @Success      201        {object}  dtos.SuccessResponse{data=dtos.CharacterResponse}
// @Failure      400        {object}  dtos.ErrorResponse
// @Failure      500        {object}  dtos.ErrorResponse
// @Router       /characters [post]
func (h *CharacterHandler) CreateCharacter(c *gin.Context) {
	var req dtos.CreateCharacterRequest
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
	resp, err := h.service.CreateCharacter(req)
	if err != nil {
		if containsStr(err.Error(), "invalid skills") {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Error: dtos.ErrorDetail{Code: "INVALID_SKILLS", Message: err.Error()},
			})
			return
		}
		h.logger.Error("Failed to create character", logging.F("error", err.Error()))
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{Code: "INTERNAL_ERROR", Message: "Failed to create character"},
		})
		return
	}
	c.JSON(http.StatusCreated, dtos.SuccessResponse{Data: resp})
}

// UpdateCharacter handles PUT /api/characters/:id
//
// @Summary      Update a character
// @Tags         characters
// @Accept       json
// @Produce      json
// @Param        id         path      string                       true  "Character UUID"
// @Param        character  body      dtos.UpdateCharacterRequest  true  "Patch payload"
// @Success      200        {object}  dtos.SuccessResponse{data=dtos.CharacterResponse}
// @Failure      400        {object}  dtos.ErrorResponse
// @Failure      404        {object}  dtos.ErrorResponse
// @Failure      500        {object}  dtos.ErrorResponse
// @Router       /characters/{id} [put]
func (h *CharacterHandler) UpdateCharacter(c *gin.Context) {
	id := c.Param("id")
	var req dtos.UpdateCharacterRequest
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
	resp, err := h.service.UpdateCharacter(id, req)
	if err != nil {
		if containsStr(err.Error(), "invalid UUID") {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Error: dtos.ErrorDetail{Code: "INVALID_ID", Message: "Invalid character ID"},
			})
			return
		}
		if containsStr(err.Error(), "invalid skills") {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Error: dtos.ErrorDetail{Code: "INVALID_SKILLS", Message: err.Error()},
			})
			return
		}
		h.logger.Error("Failed to update character", logging.F("error", err.Error()))
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{Code: "INTERNAL_ERROR", Message: "Failed to update character"},
		})
		return
	}
	if resp == nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{Code: "CHARACTER_NOT_FOUND", Message: "Character not found"},
		})
		return
	}
	c.JSON(http.StatusOK, dtos.SuccessResponse{Data: resp})
}

// DeleteCharacter handles DELETE /api/characters/:id
//
// @Summary      Delete a character
// @Tags         characters
// @Param        id   path      string  true  "Character UUID"
// @Success      204
// @Failure      400  {object}  dtos.ErrorResponse
// @Failure      404  {object}  dtos.ErrorResponse
// @Failure      500  {object}  dtos.ErrorResponse
// @Router       /characters/{id} [delete]
func (h *CharacterHandler) DeleteCharacter(c *gin.Context) {
	id := c.Param("id")
	err := h.service.DeleteCharacter(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{
				Error: dtos.ErrorDetail{Code: "CHARACTER_NOT_FOUND", Message: "Character not found"},
			})
			return
		}
		if containsStr(err.Error(), "invalid UUID") {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Error: dtos.ErrorDetail{Code: "INVALID_ID", Message: "Invalid character ID"},
			})
			return
		}
		h.logger.Error("Failed to delete character", logging.F("error", err.Error()))
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Error: dtos.ErrorDetail{Code: "INTERNAL_ERROR", Message: "Failed to delete character"},
		})
		return
	}
	c.Status(http.StatusNoContent)
}
