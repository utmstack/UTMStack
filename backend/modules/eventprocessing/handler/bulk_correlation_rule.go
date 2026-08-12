package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type BulkCorrelationRuleHandler struct {
	uc           connectors.CorrelationRuleUsecase
	tenantLister func(context.Context) ([]string, error)
}

func NewBulkCorrelationRuleHandler(uc connectors.CorrelationRuleUsecase, tenantLister func(context.Context) ([]string, error)) *BulkCorrelationRuleHandler {
	return &BulkCorrelationRuleHandler{uc: uc, tenantLister: tenantLister}
}

// @Summary     Bulk create correlation rule
// @Description Installs the same correlation rule in N tenants (each gets its own copy in its user overlay).
// @Tags        Platform / Correlation Rules
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.BulkCreateCorrelationRuleRequest true "selector + rule"
// @Success     200 {object} common_models.BulkResult
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /platform/eventprocessing/correlation-rule/bulk/create [post]
func (h *BulkCorrelationRuleHandler) Create(c *gin.Context) {
	var req dto.BulkCreateCorrelationRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantIDs, err := resolveTenants(c.Request.Context(), req.Selector, h.tenantLister)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var result common_models.BulkResult
	for _, tid := range tenantIDs {
		ctx := authz.WithTenantID(c.Request.Context(), tid)
		result.Append(tid, h.uc.Create(ctx, req.Rule))
	}
	c.JSON(http.StatusOK, result)
}

// @Summary     Bulk update correlation rule
// @Description Updates the correlation rule with matching relPath across N tenants. System-owned rules refuse per tenant.
// @Tags        Platform / Correlation Rules
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.BulkUpdateCorrelationRuleRequest true "selector + rule (includes relPath)"
// @Success     200 {object} common_models.BulkResult
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /platform/eventprocessing/correlation-rule/bulk/update [post]
func (h *BulkCorrelationRuleHandler) Update(c *gin.Context) {
	var req dto.BulkUpdateCorrelationRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantIDs, err := resolveTenants(c.Request.Context(), req.Selector, h.tenantLister)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var result common_models.BulkResult
	for _, tid := range tenantIDs {
		ctx := authz.WithTenantID(c.Request.Context(), tid)
		result.Append(tid, h.uc.Update(ctx, req.Rule))
	}
	c.JSON(http.StatusOK, result)
}

// @Summary     Bulk delete correlation rule
// @Description Deletes the correlation rule with matching relPath across N tenants. System-owned rules refuse per tenant.
// @Tags        Platform / Correlation Rules
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.BulkDeleteCorrelationRuleRequest true "selector + relPath"
// @Success     200 {object} common_models.BulkResult
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /platform/eventprocessing/correlation-rule/bulk/delete [post]
func (h *BulkCorrelationRuleHandler) Delete(c *gin.Context) {
	var req dto.BulkDeleteCorrelationRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantIDs, err := resolveTenants(c.Request.Context(), req.Selector, h.tenantLister)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var result common_models.BulkResult
	for _, tid := range tenantIDs {
		ctx := authz.WithTenantID(c.Request.Context(), tid)
		result.Append(tid, h.uc.Delete(ctx, req.RelPath))
	}
	c.JSON(http.StatusOK, result)
}

// @Summary     Bulk activate/deactivate correlation rule
// @Description Enables or disables the correlation rule per tenant.
// @Tags        Platform / Correlation Rules
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.BulkActivateCorrelationRuleRequest true "selector + relPath + active"
// @Success     200 {object} common_models.BulkResult
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /platform/eventprocessing/correlation-rule/bulk/activate [post]
func (h *BulkCorrelationRuleHandler) Activate(c *gin.Context) {
	var req dto.BulkActivateCorrelationRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantIDs, err := resolveTenants(c.Request.Context(), req.Selector, h.tenantLister)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var result common_models.BulkResult
	for _, tid := range tenantIDs {
		ctx := authz.WithTenantID(c.Request.Context(), tid)
		_, err := h.uc.SetActive(ctx, req.RelPath, req.Active)
		result.Append(tid, err)
	}
	c.JSON(http.StatusOK, result)
}
