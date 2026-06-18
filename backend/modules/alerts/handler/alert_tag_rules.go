package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/modules/alerts/dto"
)

type AlertTagRuleHandler struct {
	usecase connectors.AlertTagRuleUsecase
}

func NewAlertTagRuleHandler(uc connectors.AlertTagRuleUsecase) *AlertTagRuleHandler {
	return &AlertTagRuleHandler{usecase: uc}
}

// @Summary     Create alert tag rule
// @Tags        Alert Tag Rules
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.CreateAlertTagRuleRequest true "Alert tag rule"
// @Success     200 {object} dto.AlertTagRuleResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /alert-tag-rules [post]
func (h *AlertTagRuleHandler) Create(c *gin.Context) {
	var req dto.CreateAlertTagRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule, err := h.usecase.Create(c.Request.Context(), req)
	if err != nil {
		writeAlertError(c, err)
		return
	}
	c.JSON(http.StatusOK, rule)
}

// @Summary     Update alert tag rule
// @Tags        Alert Tag Rules
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.UpdateAlertTagRuleRequest true "Alert tag rule"
// @Success     200 {object} dto.AlertTagRuleResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /alert-tag-rules [put]
func (h *AlertTagRuleHandler) Update(c *gin.Context) {
	var req dto.UpdateAlertTagRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule, err := h.usecase.Update(c.Request.Context(), req)
	if err != nil {
		writeAlertError(c, err)
		return
	}
	c.JSON(http.StatusOK, rule)
}

// @Summary     List alert tag rules
// @Tags        Alert Tag Rules
// @Security    BearerAuth
// @Produce     json
// @Param       id             query int    false "Filter by ID"
// @Param       name           query string false "Filter by name"
// @Param       conditionField query string false "Filter by condition field"
// @Param       conditionValue query string false "Filter by condition value"
// @Param       ruleActive     query bool   false "Filter by active status"
// @Param       ruleDeleted    query bool   false "Filter by deleted status"
// @Param       tagIds         query string false "Comma-separated tag IDs"
// @Param       page           query int    false "Page (default 1)"
// @Param       size           query int    false "Page size (default 20)"
// @Success     200 {array} dto.AlertTagRuleResponse
// @Header      200 {string} X-Total-Count "Total items"
// @Failure     500 {object} map[string]string
// @Router      /alert-tag-rules [get]
func (h *AlertTagRuleHandler) List(c *gin.Context) {
	filters := parseAlertTagRuleFilters(c)
	rows, total, err := h.usecase.List(c.Request.Context(), filters)
	if err != nil {
		writeAlertError(c, err)
		return
	}
	// Return bare array + X-Total-Count (matches Java Spring Page contract).
	writePagedArray(c, rows, total)
}

// getByIDsRequest holds the repeated ?ids=1&ids=2 params from query string.
// Java: @RequestParam List<Long> ids — repeated param style.
type getByIDsRequest struct {
	IDs []int64 `form:"ids"`
}

// @Summary     Get alert tag rules by IDs
// @Tags        Alert Tag Rules
// @Security    BearerAuth
// @Produce     json
// @Param       ids query []int64 true "Rule IDs (repeated param)"
// @Success     200 {array} dto.AlertTagRuleResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /alert-tag-rules/get-by-ids [get]
func (h *AlertTagRuleHandler) GetByIDs(c *gin.Context) {
	var req getByIDsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.IDs) == 0 {
		c.JSON(http.StatusOK, []dto.AlertTagRuleResponse{})
		return
	}
	rules, err := h.usecase.GetByIDs(c.Request.Context(), req.IDs)
	if err != nil {
		writeAlertError(c, err)
		return
	}
	if rules == nil {
		rules = []dto.AlertTagRuleResponse{}
	}
	c.JSON(http.StatusOK, rules)
}

// @Summary     Get alert tag rule by ID
// @Tags        Alert Tag Rules
// @Security    BearerAuth
// @Produce     json
// @Param       id path int true "Rule ID"
// @Success     200 {object} dto.AlertTagRuleResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /alert-tag-rules/{id} [get]
func (h *AlertTagRuleHandler) GetByID(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	rule, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		// Java: ResponseUtil.wrapOrNotFound returns 200 with empty body when not found.
		if isNotFoundErr(err) {
			c.JSON(http.StatusOK, nil)
			return
		}
		writeAlertError(c, err)
		return
	}
	c.JSON(http.StatusOK, rule)
}

// @Summary     Delete alert tag rule
// @Tags        Alert Tag Rules
// @Security    BearerAuth
// @Param       id path int true "Rule ID"
// @Success     200 "Deleted"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /alert-tag-rules/{id} [delete]
func (h *AlertTagRuleHandler) Delete(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		writeAlertError(c, err)
		return
	}
	// Java returns 200 OK on delete.
	c.Status(http.StatusOK)
}

// @Summary     List active alert tag rules (internal)
// @Tags        Alert Tag Rules
// @Security    InternalKey
// @Produce     json
// @Success     200 {array} dto.ActiveAlertTagRule
// @Failure     500 {object} map[string]string
// @Router      /internal/alert-tag-rules/active [get]
func (h *AlertTagRuleHandler) ListActive(c *gin.Context) {
	rules, err := h.usecase.ListActiveResolved(c.Request.Context())
	if err != nil {
		writeAlertError(c, err)
		return
	}
	if rules == nil {
		rules = []dto.ActiveAlertTagRule{}
	}
	c.JSON(http.StatusOK, rules)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func parseAlertTagRuleFilters(c *gin.Context) dto.AlertTagRuleFilters {
	filters := dto.AlertTagRuleFilters{
		Page: queryInt(c, "page", 1),
		Size: queryInt(c, "size", 20),
	}
	if v := c.Query("id"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			id := n
			filters.ID = &id
		}
	}
	if v := c.Query("name"); v != "" {
		filters.Name = &v
	}
	if v := c.Query("conditionField"); v != "" {
		filters.ConditionField = &v
	}
	if v := c.Query("conditionValue"); v != "" {
		filters.ConditionValue = &v
	}
	if v := c.Query("ruleActive"); v != "" {
		b := v == "true"
		filters.RuleActive = &b
	}
	if v := c.Query("ruleDeleted"); v != "" {
		b := v == "true"
		filters.RuleDeleted = &b
	}
	if v := c.Query("tagIds"); v != "" {
		filters.TagIDs = parseIDs(v)
	}
	return filters
}

func parseIDs(raw string) []int64 {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	ids := make([]int64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if n, err := strconv.ParseInt(p, 10, 64); err == nil && n > 0 {
			ids = append(ids, n)
		}
	}
	return ids
}
