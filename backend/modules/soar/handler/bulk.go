package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

// BulkHandler serves platform-admin bulk endpoints for soar rules.
type BulkHandler struct {
	ruleUC       connectors.RuleUsecase
	tenantLister func(context.Context) ([]string, error)
}

func NewBulkHandler(
	ruleUC connectors.RuleUsecase,
	tenantLister func(context.Context) ([]string, error),
) *BulkHandler {
	return &BulkHandler{ruleUC: ruleUC, tenantLister: tenantLister}
}

func resolveTenants(ctx context.Context, sel common_models.BulkTenantSelector, lister func(context.Context) ([]string, error)) ([]string, error) {
	if sel.AllTenants {
		return lister(ctx)
	}
	return sel.TenantIDs, nil
}

// BulkCreateRule godoc
//
//	@Summary     Bulk create SOAR rule across tenants
//	@Tags        Platform SOAR
//	@Security    BearerAuth
//	@Accept      json
//	@Produce     json
//	@Param       body body dto.BulkCreateRuleRequest true "Request"
//	@Success     200 {object} common_models.BulkResult
//	@Router      /platform/soar/rules/bulk/create [post]
func (h *BulkHandler) BulkCreateRule(c *gin.Context) {
	var req dto.BulkCreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tids, err := resolveTenants(c.Request.Context(), req.Selector, h.tenantLister)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var result common_models.BulkResult
	for _, tid := range tids {
		ctx := authz.WithTenantID(c.Request.Context(), tid)
		_, err := h.ruleUC.Create(ctx, req.Rule, "platform-admin")
		result.Append(tid, err)
	}
	c.JSON(http.StatusOK, result)
}

// BulkUpdateRule godoc
//
//	@Summary     Bulk update SOAR rule across tenants
//	@Tags        Platform SOAR
//	@Security    BearerAuth
//	@Accept      json
//	@Produce     json
//	@Param       body body dto.BulkUpdateRuleRequest true "Request"
//	@Success     200 {object} common_models.BulkResult
//	@Router      /platform/soar/rules/bulk/update [post]
func (h *BulkHandler) BulkUpdateRule(c *gin.Context) {
	var req dto.BulkUpdateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tids, err := resolveTenants(c.Request.Context(), req.Selector, h.tenantLister)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var result common_models.BulkResult
	for _, tid := range tids {
		ctx := authz.WithTenantID(c.Request.Context(), tid)
		_, err := h.ruleUC.Update(ctx, req.RelPath, req.Rule, "platform-admin")
		result.Append(tid, err)
	}
	c.JSON(http.StatusOK, result)
}

// BulkDeleteRule godoc
//
//	@Summary     Bulk delete SOAR rule across tenants
//	@Tags        Platform SOAR
//	@Security    BearerAuth
//	@Accept      json
//	@Produce     json
//	@Param       body body dto.BulkDeleteRuleRequest true "Request"
//	@Success     200 {object} common_models.BulkResult
//	@Router      /platform/soar/rules/bulk/delete [post]
func (h *BulkHandler) BulkDeleteRule(c *gin.Context) {
	var req dto.BulkDeleteRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tids, err := resolveTenants(c.Request.Context(), req.Selector, h.tenantLister)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var result common_models.BulkResult
	for _, tid := range tids {
		ctx := authz.WithTenantID(c.Request.Context(), tid)
		result.Append(tid, h.ruleUC.Delete(ctx, req.RelPath))
	}
	c.JSON(http.StatusOK, result)
}

// BulkEnableRule godoc
//
//	@Summary     Bulk enable/disable SOAR rule across tenants
//	@Tags        Platform SOAR
//	@Security    BearerAuth
//	@Accept      json
//	@Produce     json
//	@Param       body body dto.BulkEnableRuleRequest true "Request"
//	@Success     200 {object} common_models.BulkResult
//	@Router      /platform/soar/rules/bulk/enable [post]
func (h *BulkHandler) BulkEnableRule(c *gin.Context) {
	var req dto.BulkEnableRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tids, err := resolveTenants(c.Request.Context(), req.Selector, h.tenantLister)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var result common_models.BulkResult
	for _, tid := range tids {
		ctx := authz.WithTenantID(c.Request.Context(), tid)
		result.Append(tid, h.ruleUC.SetEnabled(ctx, req.RelPath, req.Enabled))
	}
	c.JSON(http.StatusOK, result)
}
