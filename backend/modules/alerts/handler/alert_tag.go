package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/modules/alerts/dto"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
)

// ---------------------------------------------------------------------------
// Tag catalog
// ---------------------------------------------------------------------------

type AlertTagHandler struct {
	usecase connectors.AlertTagUsecase
}

func NewAlertTagHandler(uc connectors.AlertTagUsecase) *AlertTagHandler {
	return &AlertTagHandler{usecase: uc}
}

// @Summary     Create alert tag
// @Tags        Alert Tags
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.CreateAlertTagRequest true "Alert tag"
// @Success     201 {object} dto.AlertTagResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-alert-tags [post]
func (h *AlertTagHandler) Create(c *gin.Context) {
	var req dto.CreateAlertTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tag, err := h.usecase.Create(c.Request.Context(), req)
	if err != nil {
		writeAlertError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toAlertTagResponse(tag))
}

// @Summary     Update alert tag
// @Tags        Alert Tags
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.UpdateAlertTagRequest true "Alert tag"
// @Success     200 {object} dto.AlertTagResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-alert-tags [put]
func (h *AlertTagHandler) Update(c *gin.Context) {
	var req dto.UpdateAlertTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tag, err := h.usecase.Update(c.Request.Context(), req)
	if err != nil {
		writeAlertError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAlertTagResponse(tag))
}

// @Summary     List alert tags
// @Tags        Alert Tags
// @Security    BearerAuth
// @Produce     json
// @Param       tagName     query string false "Filter by tag name"
// @Param       systemOwner query bool   false "Filter by system owner"
// @Param       page        query int    false "Page (default 1)"
// @Param       size        query int    false "Page size (default 20)"
// @Success     200 {array} dto.AlertTagResponse
// @Header      200 {string} X-Total-Count "Total items"
// @Failure     500 {object} map[string]string
// @Router      /utm-alert-tags [get]
func (h *AlertTagHandler) List(c *gin.Context) {
	filters := parseAlertTagFilters(c)
	rows, total, err := h.usecase.List(c.Request.Context(), filters)
	if err != nil {
		writeAlertError(c, err)
		return
	}

	responses := make([]dto.AlertTagResponse, 0, len(rows))
	for i := range rows {
		responses = append(responses, toAlertTagResponse(&rows[i]))
	}

	// Return bare array + X-Total-Count (matches Java Spring Page contract).
	writePagedArray(c, responses, total)
}

// @Summary     Get alert tag by ID
// @Tags        Alert Tags
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Alert tag ID"
// @Success     200 {object} dto.AlertTagResponse
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-alert-tags/{id} [get]
func (h *AlertTagHandler) GetByID(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	tag, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		writeAlertError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAlertTagResponse(tag))
}

// @Summary     Delete alert tag
// @Tags        Alert Tags
// @Security    BearerAuth
// @Param       id path string true "Alert tag ID"
// @Success     200 "Deleted"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-alert-tags/{id} [delete]
func (h *AlertTagHandler) Delete(c *gin.Context) {
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

// ---------------------------------------------------------------------------
// Tagging rules
// ---------------------------------------------------------------------------

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
	audit.Record(c, audit_connectors.Event{
		Action:       "alert_tag_rule.create",
		ResourceType: "alert_tag_rule",
		ResourceID:   ruleID(rule),
		Metadata:     map[string]any{"name": req.Name, "tags": tagNames(req.Tags)},
	}, audit_domain.ALERT_TAG_RULE_CREATE_ATTEMPT, audit_domain.ALERT_TAG_RULE_CREATE_SUCCESS, err)
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
	audit.Record(c, audit_connectors.Event{
		Action:       "alert_tag_rule.update",
		ResourceType: "alert_tag_rule",
		ResourceID:   req.ID.String(),
		Metadata:     map[string]any{"name": req.Name, "tags": tagNames(req.Tags)},
	}, audit_domain.ALERT_TAG_RULE_UPDATE_ATTEMPT, audit_domain.ALERT_TAG_RULE_UPDATE_SUCCESS, err)
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
// @Param       id             query string false "Filter by ID"
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

// getByIDsRequest holds the repeated ?ids=…&ids=… params from the query string.
type getByIDsRequest struct {
	IDs []uuid.UUID `form:"ids"`
}

// @Summary     Get alert tag rules by IDs
// @Tags        Alert Tag Rules
// @Security    BearerAuth
// @Produce     json
// @Param       ids query []string true "Rule IDs (repeated param)"
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
// @Param       id path string true "Rule ID"
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
// @Param       id path string true "Rule ID"
// @Success     200 "Deleted"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /alert-tag-rules/{id} [delete]
func (h *AlertTagRuleHandler) Delete(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	err := h.usecase.Delete(c.Request.Context(), id)
	audit.Record(c, audit_connectors.Event{
		Action:       "alert_tag_rule.delete",
		ResourceType: "alert_tag_rule",
		ResourceID:   id.String(),
	}, audit_domain.ALERT_TAG_RULE_DELETE_ATTEMPT, audit_domain.ALERT_TAG_RULE_DELETE_SUCCESS, err)
	if err != nil {
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

func parseAlertTagFilters(c *gin.Context) dto.AlertTagFilters {
	filters := dto.AlertTagFilters{
		Page: queryInt(c, "page", 1),
		Size: queryInt(c, "size", 20),
	}
	if v := c.Query("tagName"); v != "" {
		filters.TagName = &v
	}
	if v := c.Query("systemOwner"); v != "" {
		b := v == "true"
		filters.SystemOwner = &b
	}
	return filters
}

func toAlertTagResponse(tag *domain.AlertTag) dto.AlertTagResponse {
	return dto.AlertTagResponse{
		ID:          tag.ID,
		TagName:     tag.TagName,
		TagColor:    tag.TagColor,
		SystemOwner: tag.SystemOwner,
	}
}

func parseAlertTagRuleFilters(c *gin.Context) dto.AlertTagRuleFilters {
	filters := dto.AlertTagRuleFilters{
		Page: queryInt(c, "page", 1),
		Size: queryInt(c, "size", 20),
	}
	if v := c.Query("id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
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

func parseIDs(v string) []uuid.UUID {
	parts := strings.Split(v, ",")
	out := make([]uuid.UUID, 0, len(parts))
	for _, p := range parts {
		if id, err := uuid.Parse(strings.TrimSpace(p)); err == nil {
			out = append(out, id)
		}
	}
	return out
}

// ruleID names the resource a create audited, which is only known once the
// rule exists — a failed create has none.
func ruleID(rule *dto.AlertTagRuleResponse) string {
	if rule == nil {
		return ""
	}
	return rule.ID.String()
}

// tagNames records what the rule applies rather than the tag ids, so the entry
// stays readable after a tag is renamed or removed.
func tagNames(tags []dto.AlertTagRuleTagRef) []string {
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		out = append(out, t.TagName)
	}
	return out
}
