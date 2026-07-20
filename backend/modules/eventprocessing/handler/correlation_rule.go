package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
)

// TODO(module-33): re-evaluate if network-scan property values are needed here.
var allowedPropertyColumns = map[string]bool{
	"rule_name":      true,
	"rule_category":  true,
	"rule_technique": true,
	"rule_adversary": true,
}

type CorrelationRuleHandler struct {
	usecase connectors.CorrelationRuleUsecase
}

func NewCorrelationRuleHandler(uc connectors.CorrelationRuleUsecase) *CorrelationRuleHandler {
	return &CorrelationRuleHandler{usecase: uc}
}

// @Summary     Create correlation rule
// @Tags        Correlation Rules
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.CreateCorrelationRuleRequest true "Correlation rule to create"
// @Success     204 "No content"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /correlation-rule [post]
func (h *CorrelationRuleHandler) Create(c *gin.Context) {
	var req dto.CreateCorrelationRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.usecase.Create(c.Request.Context(), req)
	audit.Record(c, audit_connectors.Event{Action: "correlation_rule.create"}, audit_domain.CORRELATION_RULE_CREATE_ATTEMPT, audit_domain.CORRELATION_RULE_CREATE_SUCCESS, err)
	if err != nil {
		writeCorrelationError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     Import correlation rules from YAML
// @Description Validates each uploaded rule YAML and creates the valid ones as user rules. Always 200 with a per-file verdict.
// @Tags        Correlation Rules
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.ImportCorrelationRulesRequest true "Rule YAML files"
// @Success     200 {object} dto.ImportCorrelationRulesResponse
// @Failure     400 {object} map[string]string
// @Router      /correlation-rule/import [post]
func (h *CorrelationRuleHandler) Import(c *gin.Context) {
	var req dto.ImportCorrelationRulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results := h.usecase.ImportRules(c.Request.Context(), req.Files)
	approved := 0
	for _, r := range results {
		if r.Approved {
			approved++
		}
	}
	audit.Record(c, audit_connectors.Event{Action: "correlation_rule.import"}, audit_domain.CORRELATION_RULE_CREATE_ATTEMPT, audit_domain.CORRELATION_RULE_CREATE_SUCCESS, nil)
	c.JSON(http.StatusOK, dto.ImportCorrelationRulesResponse{
		Results:  results,
		Approved: approved,
		Rejected: len(results) - approved,
	})
}

// @Summary     Update correlation rule
// @Tags        Correlation Rules
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.UpdateCorrelationRuleRequest true "Correlation rule to update"
// @Success     204 "No content"
// @Failure     400 {object} map[string]string "Validation error or system rule cannot be modified"
// @Failure     404 {object} map[string]string "Correlation rule not found"
// @Failure     500 {object} map[string]string
// @Router      /correlation-rule [put]
func (h *CorrelationRuleHandler) Update(c *gin.Context) {
	var req dto.UpdateCorrelationRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.usecase.Update(c.Request.Context(), req)
	audit.Record(c, audit_connectors.Event{Action: "correlation_rule.update"}, audit_domain.CORRELATION_RULE_UPDATE_ATTEMPT, audit_domain.CORRELATION_RULE_UPDATE_SUCCESS, err)
	if err != nil {
		if isNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "correlation rule not found"})
			return
		}
		writeCorrelationError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     Activate or deactivate a correlation rule
// @Tags        Correlation Rules
// @Security    BearerAuth
// @Produce     json
// @Param       id     query int64 true  "Correlation rule ID"
// @Param       active query bool  true  "true to activate, false to deactivate"
// @Success     200 {object} map[string]bool "changed: true if this call actually flipped the rule's state"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /correlation-rule/activate-deactivate [put]
func (h *CorrelationRuleHandler) ActivateDeactivate(c *gin.Context) {
	relPath := c.Query("relPath")
	if relPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "relPath is required"})
		return
	}

	active, err := strconv.ParseBool(c.Query("active"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid active parameter: must be true or false"})
		return
	}

	changed, err := h.usecase.SetActive(c.Request.Context(), relPath, active)
	audit.Record(c, audit_connectors.Event{Action: "correlation_rule.activate"}, audit_domain.CORRELATION_RULE_UPDATE_ATTEMPT, audit_domain.CORRELATION_RULE_UPDATE_SUCCESS, err)
	if err != nil {
		writeCorrelationError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"changed": changed})
}

// @Summary     List correlation rules by filters
// @Tags        Correlation Rules
// @Security    BearerAuth
// @Produce     json
// @Param       page                 query int      false "Page (0-based)"
// @Param       size                 query int      false "Page size"
// @Param       ruleName             query string   false "Partial match on rule name"
// @Param       ruleActive           query boolean  false "Filter by active status"
// @Param       ruleCategory         query []string false "Filter by category (repeatable)"
// @Param       ruleAdversary        query []string false "Filter by adversary (repeatable)"
// @Param       ruleTechnique        query []string false "Filter by technique (repeatable)"
// @Param       ruleConfidentiality  query []int    false "Filter by confidentiality level (repeatable)"
// @Param       ruleIntegrity        query []int    false "Filter by integrity level (repeatable)"
// @Param       ruleAvailability     query []int    false "Filter by availability level (repeatable)"
// @Param       systemOwner          query boolean  false "Filter by system ownership"
// @Param       dataTypes            query []string false "Filter by associated data type strings (repeatable)"
// @Param       initDate             query string   false "Start of last-update range (ISO-8601)"
// @Param       endDate              query string   false "End of last-update range (ISO-8601)"
// @Param       search               query string   false "General text search against rule name"
// @Success     200 {array}  dto.CorrelationRuleResponse
// @Header      200 {string} X-Total-Count "Total number of records"
// @Failure     500 {object} map[string]string
// @Router      /correlation-rule/search-by-filters [get]
func (h *CorrelationRuleHandler) List(c *gin.Context) {
	var f dto.CorrelationRuleFilters
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

// @Summary     Search distinct property values for correlation rules
// @Tags        Correlation Rules
// @Security    BearerAuth
// @Produce     json
// @Param       prop  query string true  "Property column to search (rule_name, rule_category, rule_technique, rule_adversary)"
// @Param       value query string false "Partial value to filter by"
// @Success     200 {array}  string
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /correlation-rule/search-property-values [get]
func (h *CorrelationRuleHandler) SearchPropertyValues(c *gin.Context) {
	prop := c.Query("prop")
	if !allowedPropertyColumns[prop] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or unsupported property: " + prop})
		return
	}

	value := c.Query("value")

	values, err := h.usecase.FindDistinctPropertyValues(c.Request.Context(), prop, value)
	if err != nil {
		writeCorrelationError(c, err)
		return
	}
	if values == nil {
		values = []string{}
	}
	c.JSON(http.StatusOK, values)
}

// @Summary     Get correlation rule by ID
// @Tags        Correlation Rules
// @Security    BearerAuth
// @Produce     json
// @Param       id path int true "Correlation rule ID"
// @Success     200 {object} dto.CorrelationRuleResponse
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /correlation-rule/{id} [get]
func (h *CorrelationRuleHandler) GetByID(c *gin.Context) {
	relPath := c.Query("relPath")
	if relPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "relPath is required"})
		return
	}

	rule, err := h.usecase.GetByRelPath(c.Request.Context(), relPath)
	if err != nil {
		if isNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "correlation rule not found"})
			return
		}
		writeCorrelationError(c, err)
		return
	}
	c.JSON(http.StatusOK, rule)
}

// @Summary     Delete correlation rule
// @Tags        Correlation Rules
// @Security    BearerAuth
// @Param       id path int true "Correlation rule ID"
// @Success     204 "No content"
// @Failure     400 {object} map[string]string "systemOwner rule cannot be deleted"
// @Failure     500 {object} map[string]string
// @Router      /correlation-rule/{id} [delete]
func (h *CorrelationRuleHandler) Delete(c *gin.Context) {
	relPath := c.Query("relPath")
	if relPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "relPath is required"})
		return
	}

	err := h.usecase.Delete(c.Request.Context(), relPath)
	audit.Record(c, audit_connectors.Event{Action: "correlation_rule.delete"}, audit_domain.CORRELATION_RULE_DELETE_ATTEMPT, audit_domain.CORRELATION_RULE_DELETE_SUCCESS, err)
	if err != nil {
		writeCorrelationError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
