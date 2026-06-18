package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
)

type RuleHandler struct {
	usecase connectors.RuleUsecase
}

func NewRuleHandler(uc connectors.RuleUsecase) *RuleHandler {
	return &RuleHandler{usecase: uc}
}

// @Summary     Create alert response rule (flow)
// @Description Writes the flow to the user overlay as a YAML file. Identity is the
// @Description file relPath; there is no numeric id.
// @Tags        SOAR Rules
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.CreateRuleRequest true "Flow to create"
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
	resp, err := h.usecase.Create(c.Request.Context(), req, loginFromCtx(c))
	ev := audit_connectors.Event{Action: "soar.rule.create", ResourceType: "soar_rule"}
	if resp != nil {
		ev.ResourceID = resp.RelPath
	}
	audit.Record(c, ev, audit_domain.SOAR_RULE_CREATE_ATTEMPT, audit_domain.SOAR_RULE_CREATE_SUCCESS, err)
	if err != nil {
		writeARRError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Update alert response rule (flow)
// @Tags        SOAR Rules
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       relPath path string true "Flow relPath"
// @Param       input body dto.UpdateRuleRequest true "Flow to update"
// @Success     200 {object} dto.RuleResponse
// @Failure     400 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /soar/rules/{relPath} [put]
func (h *RuleHandler) Update(c *gin.Context) {
	relPath := c.Param("relPath")
	var req dto.UpdateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.usecase.Update(c.Request.Context(), relPath, req, loginFromCtx(c))
	audit.Record(c, audit_connectors.Event{Action: "soar.rule.update", ResourceType: "soar_rule", ResourceID: relPath},
		audit_domain.SOAR_RULE_UPDATE_ATTEMPT, audit_domain.SOAR_RULE_UPDATE_SUCCESS, err)
	if err != nil {
		writeARRError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     List alert response rules (flows)
// @Tags        SOAR Rules
// @Security    BearerAuth
// @Produce     json
// @Param       page                 query int    false "Page number (0-based)"
// @Param       size                 query int    false "Page size"
// @Param       name.contains        query string false "Filter by name (substring)"
// @Param       active.equals        query bool   false "Filter by active flag"
// @Param       agentPlatform.equals query string false "Filter by agent platform"
// @Param       systemOwner.equals   query bool   false "Filter by system ownership"
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
	writePagedArray(c, result.Items, result.Total)
}

// @Summary     Get alert response rule (flow) by relPath
// @Tags        SOAR Rules
// @Security    BearerAuth
// @Produce     json
// @Param       relPath path string true "Flow relPath"
// @Success     200 {object} dto.RuleResponse
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /soar/rules/{relPath} [get]
func (h *RuleHandler) Get(c *gin.Context) {
	resp, err := h.usecase.Get(c.Request.Context(), c.Param("relPath"))
	if err != nil {
		writeARRError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Delete alert response rule (flow) by relPath
// @Tags        SOAR Rules
// @Security    BearerAuth
// @Param       relPath path string true "Flow relPath"
// @Success     204 "Deleted"
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /soar/rules/{relPath} [delete]
func (h *RuleHandler) Delete(c *gin.Context) {
	relPath := c.Param("relPath")
	err := h.usecase.Delete(c.Request.Context(), relPath)
	audit.Record(c, audit_connectors.Event{Action: "soar.rule.delete", ResourceType: "soar_rule", ResourceID: relPath},
		audit_domain.SOAR_RULE_UPDATE_ATTEMPT, audit_domain.SOAR_RULE_UPDATE_SUCCESS, err)
	if err != nil {
		writeARRError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     Enable/disable an alert response rule (flow)
// @Description Toggles the .disabled file suffix. Allowed on system flows too.
// @Tags        SOAR Rules
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       relPath path string true "Flow relPath"
// @Param       input body dto.ToggleRuleRequest true "Enabled state"
// @Success     204 "Updated"
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /soar/rules/{relPath}/enabled [put]
func (h *RuleHandler) SetEnabled(c *gin.Context) {
	relPath := c.Param("relPath")
	var req dto.ToggleRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.usecase.SetEnabled(c.Request.Context(), relPath, req.Enabled)
	audit.Record(c, audit_connectors.Event{Action: "soar.rule.toggle", ResourceType: "soar_rule", ResourceID: relPath},
		audit_domain.SOAR_RULE_UPDATE_ATTEMPT, audit_domain.SOAR_RULE_UPDATE_SUCCESS, err)
	if err != nil {
		writeARRError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
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
