package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
)

// Regex patterns are read-only: a shared vocabulary seeded by the pipeline
// bootstrap and referenced from filter YAMLs as {{.name}}. There is no create,
// update or delete surface, so nothing here records an audit event.
type RegexPatternHandler struct {
	usecase connectors.RegexPatternUsecase
}

func NewRegexPatternHandler(uc connectors.RegexPatternUsecase) *RegexPatternHandler {
	return &RegexPatternHandler{usecase: uc}
}

// @Summary     List regex patterns
// @Tags        Event Processing
// @Security    BearerAuth
// @Produce     json
// @Param       page   query int    false "Page (0-based)"
// @Param       size   query int    false "Page size"
// @Param       search query string false "Partial match on patternId"
// @Success     200 {array}  dto.RegexPatternResponse
// @Header      200 {string} X-Total-Count "Total records"
// @Failure     500 {object} map[string]string
// @Router      /eventprocessing/regex-pattern [get]
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

// @Summary     Get regex pattern by patternId
// @Tags        Event Processing
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Pattern ID (e.g. ipv4)"
// @Success     200 {object} dto.RegexPatternResponse
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /eventprocessing/regex-pattern/{id} [get]
func (h *RegexPatternHandler) GetByID(c *gin.Context) {
	patternID := c.Param("id")
	if patternID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "patternId is required"})
		return
	}
	result, err := h.usecase.GetByID(c.Request.Context(), patternID)
	if err != nil {
		writeCorrelationError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
