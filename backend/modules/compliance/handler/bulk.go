package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type BulkComplianceHandler struct {
	uc           connectors.FrameworkUsecase
	tenantLister func(context.Context) ([]string, error)
}

func NewBulkComplianceHandler(uc connectors.FrameworkUsecase, lister func(context.Context) ([]string, error)) *BulkComplianceHandler {
	return &BulkComplianceHandler{uc: uc, tenantLister: lister}
}

func resolveTenants(ctx context.Context, sel common_models.BulkTenantSelector, lister func(context.Context) ([]string, error)) ([]string, error) {
	if sel.AllTenants {
		return lister(ctx)
	}
	return sel.TenantIDs, nil
}

// @Summary     Bulk create framework across tenants
// @Tags        Platform Compliance
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.BulkCreateFrameworkRequest true "Bulk create framework"
// @Success     200 {object} common_models.BulkResult
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /platform/compliance/frameworks/bulk/create [post]
func (h *BulkComplianceHandler) CreateFramework(c *gin.Context) {
	var req dto.BulkCreateFrameworkRequest
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
		_, opErr := h.uc.CreateFramework(ctx, req.Framework)
		result.Append(tid, opErr)
		if opErr == nil {
			audit.Record(c, audit_connectors.Event{
				Action:       "bulk.compliance.framework.create",
				ResourceType: "compliance_framework",
				ResourceID:   req.Framework.Key,
				Metadata:     map[string]any{"tenantId": tid},
			}, audit_domain.COMPLIANCE_FRAMEWORK_CREATE_ATTEMPT, audit_domain.COMPLIANCE_FRAMEWORK_CREATE_SUCCESS, nil)
		}
	}
	c.JSON(http.StatusOK, result)
}

// @Summary     Bulk update framework across tenants
// @Tags        Platform Compliance
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.BulkUpdateFrameworkRequest true "Bulk update framework"
// @Success     200 {object} common_models.BulkResult
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /platform/compliance/frameworks/bulk/update [post]
func (h *BulkComplianceHandler) UpdateFramework(c *gin.Context) {
	var req dto.BulkUpdateFrameworkRequest
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
		_, opErr := h.uc.UpdateFramework(ctx, req.Framework)
		result.Append(tid, opErr)
		if opErr == nil {
			audit.Record(c, audit_connectors.Event{
				Action:       "bulk.compliance.framework.update",
				ResourceType: "compliance_framework",
				ResourceID:   req.Framework.Key,
				Metadata:     map[string]any{"tenantId": tid},
			}, audit_domain.COMPLIANCE_FRAMEWORK_UPDATE_ATTEMPT, audit_domain.COMPLIANCE_FRAMEWORK_UPDATE_SUCCESS, nil)
		}
	}
	c.JSON(http.StatusOK, result)
}

// @Summary     Bulk delete framework across tenants
// @Tags        Platform Compliance
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.BulkDeleteFrameworkRequest true "Bulk delete framework"
// @Success     200 {object} common_models.BulkResult
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /platform/compliance/frameworks/bulk/delete [post]
func (h *BulkComplianceHandler) DeleteFramework(c *gin.Context) {
	var req dto.BulkDeleteFrameworkRequest
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
		opErr := h.uc.DeleteFramework(ctx, req.FrameworkKey)
		result.Append(tid, opErr)
		if opErr == nil {
			audit.Record(c, audit_connectors.Event{
				Action:       "bulk.compliance.framework.delete",
				ResourceType: "compliance_framework",
				ResourceID:   req.FrameworkKey,
				Metadata:     map[string]any{"tenantId": tid},
			}, audit_domain.COMPLIANCE_FRAMEWORK_DELETE_ATTEMPT, audit_domain.COMPLIANCE_FRAMEWORK_DELETE_SUCCESS, nil)
		}
	}
	c.JSON(http.StatusOK, result)
}

// @Summary     Bulk create control across tenants
// @Tags        Platform Compliance
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.BulkCreateControlRequest true "Bulk create control"
// @Success     200 {object} common_models.BulkResult
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /platform/compliance/controls/bulk/create [post]
func (h *BulkComplianceHandler) CreateControl(c *gin.Context) {
	var req dto.BulkCreateControlRequest
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
		_, opErr := h.uc.CreateControl(ctx, req.Control)
		result.Append(tid, opErr)
		if opErr == nil {
			audit.Record(c, audit_connectors.Event{
				Action:       "bulk.compliance.control.create",
				ResourceType: "compliance_control",
				ResourceID:   req.Control.ID,
				Metadata:     map[string]any{"tenantId": tid},
			}, audit_domain.COMPLIANCE_CONTROL_CREATE_ATTEMPT, audit_domain.COMPLIANCE_CONTROL_CREATE_SUCCESS, nil)
		}
	}
	c.JSON(http.StatusOK, result)
}

// @Summary     Bulk update control across tenants
// @Tags        Platform Compliance
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.BulkUpdateControlRequest true "Bulk update control"
// @Success     200 {object} common_models.BulkResult
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /platform/compliance/controls/bulk/update [post]
func (h *BulkComplianceHandler) UpdateControl(c *gin.Context) {
	var req dto.BulkUpdateControlRequest
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
		_, opErr := h.uc.UpdateControl(ctx, req.Control)
		result.Append(tid, opErr)
		if opErr == nil {
			audit.Record(c, audit_connectors.Event{
				Action:       "bulk.compliance.control.update",
				ResourceType: "compliance_control",
				ResourceID:   req.Control.ID,
				Metadata:     map[string]any{"tenantId": tid},
			}, audit_domain.COMPLIANCE_CONTROL_UPDATE_ATTEMPT, audit_domain.COMPLIANCE_CONTROL_UPDATE_SUCCESS, nil)
		}
	}
	c.JSON(http.StatusOK, result)
}

// @Summary     Bulk delete control across tenants
// @Tags        Platform Compliance
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.BulkDeleteControlRequest true "Bulk delete control"
// @Success     200 {object} common_models.BulkResult
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /platform/compliance/controls/bulk/delete [post]
func (h *BulkComplianceHandler) DeleteControl(c *gin.Context) {
	var req dto.BulkDeleteControlRequest
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
		opErr := h.uc.DeleteControl(ctx, req.ControlID)
		result.Append(tid, opErr)
		if opErr == nil {
			audit.Record(c, audit_connectors.Event{
				Action:       "bulk.compliance.control.delete",
				ResourceType: "compliance_control",
				ResourceID:   req.ControlID,
				Metadata:     map[string]any{"tenantId": tid},
			}, audit_domain.COMPLIANCE_CONTROL_DELETE_ATTEMPT, audit_domain.COMPLIANCE_CONTROL_DELETE_SUCCESS, nil)
		}
	}
	c.JSON(http.StatusOK, result)
}
