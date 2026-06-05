package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/correlation/connectors"
	"github.com/utmstack/utmstack/backend/modules/correlation/dto"
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

	if err := h.usecase.Create(c.Request.Context(), req); err != nil {
		writeCorrelationError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
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

	if err := h.usecase.Update(c.Request.Context(), req); err != nil {
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
// @Success     204 "No content"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /correlation-rule/activate-deactivate [put]
func (h *CorrelationRuleHandler) ActivateDeactivate(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id parameter"})
		return
	}

	activeStr := c.Query("active")
	active, err := strconv.ParseBool(activeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid active parameter: must be true or false"})
		return
	}

	if err := h.usecase.ActivateDeactivate(c.Request.Context(), id, active); err != nil {
		writeCorrelationError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
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
	id, ok := pathInt64(c, "id")
	if !ok {
		return
	}

	rule, err := h.usecase.GetByID(c.Request.Context(), id)
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
