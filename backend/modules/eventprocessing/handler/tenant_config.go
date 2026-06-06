package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
)

type TenantConfigHandler struct {
	usecase connectors.TenantConfigUsecase
}

func NewTenantConfigHandler(uc connectors.TenantConfigUsecase) *TenantConfigHandler {
	return &TenantConfigHandler{usecase: uc}
}

// @Summary     Create asset (tenant config)
// @Tags        Event Processing
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.CreateTenantConfigRequest true "assetName + CIA + lists"
// @Success     200 {object} dto.TenantConfigResponse
// @Failure     400 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /eventprocessing/tenant-config [post]
func (h *TenantConfigHandler) Create(c *gin.Context) {
	var req dto.CreateTenantConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.usecase.Create(c.Request.Context(), req)
	audit.Record(c, audit_connectors.Event{Action: "tenant_config.create", ResourceType: "tenant_config", ResourceID: req.AssetName},
		audit_domain.TENANT_CONFIG_CREATE_ATTEMPT, audit_domain.TENANT_CONFIG_CREATE_SUCCESS, err)
	if err != nil {
		writeCorrelationError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Update asset (tenant config)
// @Tags        Event Processing
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.UpdateTenantConfigRequest true "assetName + CIA + lists"
// @Success     200 {object} dto.TenantConfigResponse
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /eventprocessing/tenant-config [put]
func (h *TenantConfigHandler) Update(c *gin.Context) {
	var req dto.UpdateTenantConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.usecase.Update(c.Request.Context(), req)
	audit.Record(c, audit_connectors.Event{Action: "tenant_config.update", ResourceType: "tenant_config", ResourceID: req.AssetName},
		audit_domain.TENANT_CONFIG_UPDATE_ATTEMPT, audit_domain.TENANT_CONFIG_UPDATE_SUCCESS, err)
	if err != nil {
		writeCorrelationError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     List assets (tenant configs)
// @Tags        Event Processing
// @Security    BearerAuth
// @Produce     json
// @Param       page   query int    false "Page (0-based)"
// @Param       size   query int    false "Page size"
// @Param       search query string false "Partial match on assetName"
// @Success     200 {array}  dto.TenantConfigResponse
// @Header      200 {string} X-Total-Count "Total records"
// @Failure     500 {object} map[string]string
// @Router      /eventprocessing/tenant-config [get]
func (h *TenantConfigHandler) List(c *gin.Context) {
	var f dto.TenantConfigFilters
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

// @Summary     Get asset by name
// @Tags        Event Processing
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Asset name"
// @Success     200 {object} dto.TenantConfigResponse
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /eventprocessing/tenant-config/{id} [get]
func (h *TenantConfigHandler) GetByID(c *gin.Context) {
	assetName := c.Param("id")
	if assetName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "assetName is required"})
		return
	}
	result, err := h.usecase.GetByID(c.Request.Context(), assetName)
	if err != nil {
		writeCorrelationError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// @Summary     Delete asset (tenant config)
// @Tags        Event Processing
// @Security    BearerAuth
// @Param       id path string true "Asset name"
// @Success     200 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /eventprocessing/tenant-config/{id} [delete]
func (h *TenantConfigHandler) Delete(c *gin.Context) {
	assetName := c.Param("id")
	if assetName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "assetName is required"})
		return
	}
	err := h.usecase.Delete(c.Request.Context(), assetName)
	audit.Record(c, audit_connectors.Event{Action: "tenant_config.delete", ResourceType: "tenant_config", ResourceID: assetName},
		audit_domain.TENANT_CONFIG_DELETE_ATTEMPT, audit_domain.TENANT_CONFIG_DELETE_SUCCESS, err)
	if err != nil {
		writeCorrelationError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
