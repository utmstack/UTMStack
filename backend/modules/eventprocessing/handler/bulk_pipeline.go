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

type BulkPipelineHandler struct {
	uc           connectors.PipelineUsecase
	tenantLister func(context.Context) ([]string, error)
}

func NewBulkPipelineHandler(uc connectors.PipelineUsecase, tenantLister func(context.Context) ([]string, error)) *BulkPipelineHandler {
	return &BulkPipelineHandler{uc: uc, tenantLister: tenantLister}
}

func resolveTenants(ctx context.Context, sel common_models.BulkTenantSelector, lister func(context.Context) ([]string, error)) ([]string, error) {
	if sel.AllTenants {
		return lister(ctx)
	}
	return sel.TenantIDs, nil
}

// @Summary     Bulk create pipeline
// @Description Installs the same pipeline in N tenants (each gets a copy in its user overlay).
// @Tags        Platform / Pipelines
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.BulkCreatePipelineRequest true "selector + relPath + content"
// @Success     200 {object} common_models.BulkResult
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /platform/eventprocessing/pipelines/bulk/create [post]
func (h *BulkPipelineHandler) Create(c *gin.Context) {
	var req dto.BulkCreatePipelineRequest
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
		_, err := h.uc.Create(ctx, dto.CreatePipelineRequest{RelPath: req.RelPath, Content: req.Content})
		result.Append(tid, err)
	}
	c.JSON(http.StatusOK, result)
}

// @Summary     Bulk update pipeline
// @Description Updates the pipeline with matching relPath across N tenants.
// @Tags        Platform / Pipelines
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.BulkUpdatePipelineRequest true "selector + relPath + content"
// @Success     200 {object} common_models.BulkResult
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /platform/eventprocessing/pipelines/bulk/update [post]
func (h *BulkPipelineHandler) Update(c *gin.Context) {
	var req dto.BulkUpdatePipelineRequest
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
		_, err := h.uc.Update(ctx, dto.UpdatePipelineRequest{RelPath: req.RelPath, Content: req.Content})
		result.Append(tid, err)
	}
	c.JSON(http.StatusOK, result)
}

// @Summary     Bulk delete pipeline
// @Description Deletes the pipeline with matching relPath across N tenants. System-pipeline guard preserved.
// @Tags        Platform / Pipelines
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.BulkDeletePipelineRequest true "selector + relPath"
// @Success     200 {object} common_models.BulkResult
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /platform/eventprocessing/pipelines/bulk/delete [post]
func (h *BulkPipelineHandler) Delete(c *gin.Context) {
	var req dto.BulkDeletePipelineRequest
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
		err := h.uc.Delete(ctx, req.RelPath)
		result.Append(tid, err)
	}
	c.JSON(http.StatusOK, result)
}

// @Summary     Bulk activate/deactivate pipeline
// @Description Enables or disables a pipeline per tenant.
// @Tags        Platform / Pipelines
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.BulkActivatePipelineRequest true "selector + relPath + active"
// @Success     200 {object} common_models.BulkResult
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /platform/eventprocessing/pipelines/bulk/activate [post]
func (h *BulkPipelineHandler) Activate(c *gin.Context) {
	var req dto.BulkActivatePipelineRequest
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
		err := h.uc.SetActive(ctx, req.RelPath, req.Active)
		result.Append(tid, err)
	}
	c.JSON(http.StatusOK, result)
}
