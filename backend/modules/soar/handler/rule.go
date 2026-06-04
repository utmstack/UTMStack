package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
)

type RuleHandler struct {
	usecase connectors.RuleUsecase
}

func NewRuleHandler(uc connectors.RuleUsecase) *RuleHandler {
	return &RuleHandler{usecase: uc}
}

// @Summary     Create alert response rule
// @Tags        SOAR Rules
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.CreateRuleRequest true "Rule to create"
// @Success     200 {object} dto.RuleResponse
// @Failure     400 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /soar/rules [post]
func (h *RuleHandler) Create(c *gin.Context) {
	var req dto.CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Reject if caller supplied an id — matches Java: if (dto.getId() != null) return BadRequest (FIX-6)
	if req.ID != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": domain.ErrIDMustBeAbsent.Error()})
		return
	}
	resp, err := h.usecase.Create(c.Request.Context(), req, loginFromCtx(c))
	ev := audit_connectors.Event{Action: "soar.rule.create", ResourceType: "soar_rule"}
	if resp != nil {
		ev.ResourceID = strconv.FormatInt(resp.ID, 10)
	}
	audit.Record(c, ev, audit_domain.SOAR_RULE_CREATE_ATTEMPT, audit_domain.SOAR_RULE_CREATE_SUCCESS, err)
	if err != nil {
		writeARRError(c, err)
		return
	}
	// Java returns 200 OK on create (ResponseEntity.ok()) — FIX-8
	c.JSON(http.StatusOK, resp)
}

// @Summary     Update alert response rule
// @Tags        SOAR Rules
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.UpdateRuleRequest true "Rule to update"
// @Success     200 {object} dto.RuleResponse
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /soar/rules [put]
func (h *RuleHandler) Update(c *gin.Context) {
	var req dto.UpdateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.ID == nil || *req.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": domain.ErrIDRequired.Error()})
		return
	}
	resp, err := h.usecase.Update(c.Request.Context(), req, loginFromCtx(c))
	audit.Record(c, audit_connectors.Event{Action: "soar.rule.update", ResourceType: "soar_rule", ResourceID: strconv.FormatInt(*req.ID, 10)},
		audit_domain.SOAR_RULE_UPDATE_ATTEMPT, audit_domain.SOAR_RULE_UPDATE_SUCCESS, err)
	if err != nil {
		writeARRError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     List alert response rules
// @Tags        SOAR Rules
// @Security    BearerAuth
// @Produce     json
// @Param       page                 query int    false "Page number (0-based)"
// @Param       size                 query int    false "Page size"
// @Param       name.contains        query string false "Filter by name (substring)"
// @Param       active.equals        query bool   false "Filter by active flag"
// @Param       agentPlatform.equals query string false "Filter by agent platform"
// @Param       createdBy.equals     query string false "Filter by creator login"
// @Success     200 {array}  dto.RuleResponse
// @Header      200 {integer} X-Total-Count "Total number of records"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /soar/rules [get]
func (h *RuleHandler) List(c *gin.Context) {
	var f dto.RuleFilters
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.usecase.List(c.Request.Context(), f)
	if err != nil {
		writeARRError(c, err)
		return
	}
	// Return bare array + X-Total-Count header (FIX-7)
	writePagedArray(c, result.Items, result.Total)
}

// @Summary     Get alert response rule by ID
// @Tags        SOAR Rules
// @Security    BearerAuth
// @Produce     json
// @Param       id path int true "Rule ID"
// @Success     200 {object} dto.RuleResponse
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /soar/rules/{id} [get]
func (h *RuleHandler) GetByID(c *gin.Context) {
	id, ok := pathInt64(c, "id")
	if !ok {
		return
	}
	resp, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		writeARRError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Resolve filter values for alert response rules
// @Tags        SOAR Rules
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} dto.ResolveFilterValuesResponse
// @Failure     500 {object} map[string]string
// @Router      /soar/rules/resolve-filter-values [get]
func (h *RuleHandler) ResolveFilterValues(c *gin.Context) {
	resp, err := h.usecase.ResolveFilterValues(c.Request.Context())
	if err != nil {
		writeARRError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
