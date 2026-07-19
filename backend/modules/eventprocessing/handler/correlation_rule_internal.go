package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
)

// @Summary     Deactivate a correlation rule by name (internal)
// @Tags        Correlation Rules
// @Accept      json
// @Produce     json
// @Param       input body dto.InternalDeactivateRuleRequest true "Rule to deactivate"
// @Success     200 {object} dto.InternalDeactivateRuleResponse
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /eventprocessing/internal/correlation-rule/deactivate [put]
func (h *CorrelationRuleHandler) InternalDeactivate(c *gin.Context) {
	var req dto.InternalDeactivateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// RuleName filtering in List is a case-insensitive partial match, so the
	// exact (case-insensitive) match is re-applied here to avoid disabling
	// unrelated rules whose name merely contains the requested substring.
	result, err := h.usecase.List(ctx, dto.CorrelationRuleFilters{RuleName: req.RuleName})
	if err != nil {
		writeCorrelationError(c, err)
		return
	}

	var matches []dto.CorrelationRuleResponse
	for _, r := range result.Items {
		if strings.EqualFold(r.RuleName, req.RuleName) {
			matches = append(matches, r)
		}
	}
	if len(matches) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "correlation rule not found"})
		return
	}

	// Two or more rules sharing an exact display name is an accepted edge
	// case (see design's Open Questions) — disable every exact match.
	changed := false
	for _, r := range matches {
		if !r.RuleActive {
			continue
		}
		if err := h.usecase.SetActive(ctx, r.RelPath, false); err != nil {
			writeCorrelationError(c, err)
			return
		}
		changed = true
	}

	c.JSON(http.StatusOK, dto.InternalDeactivateRuleResponse{Changed: changed})
}
