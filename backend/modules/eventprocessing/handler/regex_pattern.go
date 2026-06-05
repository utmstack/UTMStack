package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/correlation/connectors"
	"github.com/utmstack/utmstack/backend/modules/correlation/dto"
)

type RegexPatternHandler struct {
	usecase connectors.RegexPatternUsecase
}

func NewRegexPatternHandler(uc connectors.RegexPatternUsecase) *RegexPatternHandler {
	return &RegexPatternHandler{usecase: uc}
}

// @Summary     Create regex pattern
// @Tags        Regex Patterns
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.CreateRegexPatternRequest true "Regex pattern to create"
// @Success     204 "No content"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /regex-pattern [post]
func (h *RegexPatternHandler) Create(c *gin.Context) {
	var req dto.CreateRegexPatternRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := h.usecase.Create(c.Request.Context(), req); err != nil {
		writeCorrelationError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     Update regex pattern
// @Tags        Regex Patterns
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.UpdateRegexPatternRequest true "Regex pattern to update"
// @Success     204 "No content"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /regex-pattern [put]
func (h *RegexPatternHandler) Update(c *gin.Context) {
	var req dto.UpdateRegexPatternRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := h.usecase.Update(c.Request.Context(), req); err != nil {
		writeCorrelationError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     List regex patterns
// @Tags        Regex Patterns
// @Security    BearerAuth
// @Produce     json
// @Param       page   query int    false "Page (0-based)"
// @Param       size   query int    false "Page size"
// @Param       search query string false "Partial match on pattern ID or description"
// @Success     200 {array}  dto.RegexPatternResponse
// @Header      200 {string} X-Total-Count "Total number of records"
// @Failure     500 {object} map[string]string
// @Router      /regex-pattern [get]
func (h *RegexPatternHandler) List(c *gin.Context) {
	var f dto.RegexPatternFilters
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.usecase.List(c.Request.Context(), f)
	if err != nil {
		writeCorrelationError(c, err)
		return
	}
	writePagedArray(c, result.Items, result.Total)
}

// @Summary     Get regex pattern by ID
// @Tags        Regex Patterns
// @Security    BearerAuth
// @Produce     json
// @Param       id path int true "Regex pattern ID"
// @Success     200 {object} dto.RegexPatternResponse
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /regex-pattern/{id} [get]
func (h *RegexPatternHandler) GetByID(c *gin.Context) {
	id, ok := pathInt64(c, "id")
	if !ok {
		return
	}
	result, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		if isNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		writeCorrelationError(c, err)
		return
	}
	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// @Summary     Delete regex pattern
// @Tags        Regex Patterns
// @Security    BearerAuth
// @Param       id path int true "Regex pattern ID"
// @Success     204 "No content"
// @Failure     400 {object} map[string]string "systemOwner pattern cannot be deleted"
// @Failure     500 {object} map[string]string
// @Router      /regex-pattern/{id} [delete]
func (h *RegexPatternHandler) Delete(c *gin.Context) {
	id, ok := pathInt64(c, "id")
	if !ok {
		return
	}
	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		writeCorrelationError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
