package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
)

type RegexPatternHandler struct {
	usecase connectors.RegexPatternUsecase
}

func NewRegexPatternHandler(uc connectors.RegexPatternUsecase) *RegexPatternHandler {
	return &RegexPatternHandler{usecase: uc}
}

// @Summary     Create regex pattern
// @Tags        Event Processing
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.CreateRegexPatternRequest true "patternId + definition"
// @Success     200 {object} dto.RegexPatternResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /eventprocessing/regex-pattern [post]
func (h *RegexPatternHandler) Create(c *gin.Context) {
	var req dto.CreateRegexPatternRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.usecase.Create(c.Request.Context(), req)
	audit.Record(c, audit_connectors.Event{Action: "regex_pattern.create", ResourceType: "regex_pattern", ResourceID: req.PatternID},
		audit_domain.REGEX_PATTERN_CREATE_ATTEMPT, audit_domain.REGEX_PATTERN_CREATE_SUCCESS, err)
	if err != nil {
		writeCorrelationError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Update regex pattern
// @Tags        Event Processing
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.UpdateRegexPatternRequest true "patternId + new definition"
// @Success     200 {object} dto.RegexPatternResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /eventprocessing/regex-pattern [put]
func (h *RegexPatternHandler) Update(c *gin.Context) {
	var req dto.UpdateRegexPatternRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.usecase.Update(c.Request.Context(), req)
	audit.Record(c, audit_connectors.Event{Action: "regex_pattern.update", ResourceType: "regex_pattern", ResourceID: req.PatternID},
		audit_domain.REGEX_PATTERN_UPDATE_ATTEMPT, audit_domain.REGEX_PATTERN_UPDATE_SUCCESS, err)
	if err != nil {
		writeCorrelationError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
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

// @Summary     Delete regex pattern
// @Tags        Event Processing
// @Security    BearerAuth
// @Param       id path string true "Pattern ID (system patterns cannot be deleted)"
// @Success     200 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /eventprocessing/regex-pattern/{id} [delete]
func (h *RegexPatternHandler) Delete(c *gin.Context) {
	patternID := c.Param("id")
	if patternID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "patternId is required"})
		return
	}
	err := h.usecase.Delete(c.Request.Context(), patternID)
	audit.Record(c, audit_connectors.Event{Action: "regex_pattern.delete", ResourceType: "regex_pattern", ResourceID: patternID},
		audit_domain.REGEX_PATTERN_DELETE_ATTEMPT, audit_domain.REGEX_PATTERN_DELETE_SUCCESS, err)
	if err != nil {
		writeCorrelationError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
