package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type BulkIDPHandler struct {
	uc           connectors.IdentityProviderUsecase
	tenantLister func(context.Context) ([]string, error)
}

func NewBulkIDPHandler(uc connectors.IdentityProviderUsecase, tenantLister func(context.Context) ([]string, error)) *BulkIDPHandler {
	return &BulkIDPHandler{uc: uc, tenantLister: tenantLister}
}

// ponytail: resolveTenants duplicated from eventprocessing — same package boundary, not worth a shared pkg
func resolveIDPTenants(ctx context.Context, sel common_models.BulkTenantSelector, lister func(context.Context) ([]string, error)) ([]string, error) {
	if sel.AllTenants {
		return lister(ctx)
	}
	return sel.TenantIDs, nil
}

// @Summary     Bulk create identity provider
// @Description Creates the same IdP config in N tenants.
// @Tags        Platform / Identity Providers
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.BulkCreateIDPRequest true "selector + provider config"
// @Success     200 {object} common_models.BulkResult
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /platform/identity-providers/bulk/create [post]
func (h *BulkIDPHandler) Create(c *gin.Context) {
	var req dto.BulkCreateIDPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantIDs, err := resolveIDPTenants(c.Request.Context(), req.Selector, h.tenantLister)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var result common_models.BulkResult
	for _, tid := range tenantIDs {
		ctx := authz.WithTenantID(c.Request.Context(), tid)
		_, err := h.uc.Create(ctx, req.Provider)
		result.Append(tid, err)
	}
	c.JSON(http.StatusOK, result)
}

// @Summary     Bulk update identity provider
// @Description Updates an IdP config across N tenants.
// @Tags        Platform / Identity Providers
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.BulkUpdateIDPRequest true "selector + provider config"
// @Success     200 {object} common_models.BulkResult
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /platform/identity-providers/bulk/update [post]
func (h *BulkIDPHandler) Update(c *gin.Context) {
	var req dto.BulkUpdateIDPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantIDs, err := resolveIDPTenants(c.Request.Context(), req.Selector, h.tenantLister)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var result common_models.BulkResult
	for _, tid := range tenantIDs {
		ctx := authz.WithTenantID(c.Request.Context(), tid)
		_, err := h.uc.Update(ctx, req.Provider)
		result.Append(tid, err)
	}
	c.JSON(http.StatusOK, result)
}

// @Summary     Bulk delete identity provider
// @Description Deletes an IdP by ID across N tenants.
// @Tags        Platform / Identity Providers
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.BulkDeleteIDPRequest true "selector + providerId"
// @Success     200 {object} common_models.BulkResult
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /platform/identity-providers/bulk/delete [post]
func (h *BulkIDPHandler) Delete(c *gin.Context) {
	var req dto.BulkDeleteIDPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	providerID, err := uuid.Parse(req.ProviderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid providerId"})
		return
	}
	tenantIDs, err := resolveIDPTenants(c.Request.Context(), req.Selector, h.tenantLister)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var result common_models.BulkResult
	for _, tid := range tenantIDs {
		ctx := authz.WithTenantID(c.Request.Context(), tid)
		err := h.uc.Delete(ctx, providerID)
		result.Append(tid, err)
	}
	c.JSON(http.StatusOK, result)
}
