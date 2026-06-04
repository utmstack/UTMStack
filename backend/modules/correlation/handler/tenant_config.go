package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/correlation/connectors"
	"github.com/utmstack/utmstack/backend/modules/correlation/dto"
)

type TenantConfigHandler struct {
	usecase connectors.TenantConfigUsecase
}

func NewTenantConfigHandler(uc connectors.TenantConfigUsecase) *TenantConfigHandler {
	return &TenantConfigHandler{usecase: uc}
}

// @Summary     Create tenant config
// @Tags        Tenant Config
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.CreateTenantConfigRequest true "Tenant config to create"
// @Success     204 "No content"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /tenant-config [post]
func (h *TenantConfigHandler) Create(c *gin.Context) {
	var req dto.CreateTenantConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := h.usecase.Create(c.Request.Context(), req); err != nil {
		writeCorrelationError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     Update tenant config
// @Tags        Tenant Config
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.UpdateTenantConfigRequest true "Tenant config to update"
// @Success     204 "No content"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /tenant-config [put]
func (h *TenantConfigHandler) Update(c *gin.Context) {
	var req dto.UpdateTenantConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := h.usecase.Update(c.Request.Context(), req); err != nil {
		writeCorrelationError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     List tenant configs
// @Tags        Tenant Config
// @Security    BearerAuth
// @Produce     json
// @Param       page query int false "Page (0-based)"
// @Param       size query int false "Page size"
// @Success     200 {array}  dto.TenantConfigResponse
// @Header      200 {string} X-Total-Count "Total number of records"
// @Failure     500 {object} map[string]string
// @Router      /tenant-config [get]
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

// @Summary     Get tenant config by ID
// @Tags        Tenant Config
// @Security    BearerAuth
// @Produce     json
// @Param       id path int true "Tenant config ID"
// @Success     200 {object} dto.TenantConfigResponse
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /tenant-config/{id} [get]
func (h *TenantConfigHandler) GetByID(c *gin.Context) {
	id, ok := pathInt64(c, "id")
	if !ok {
		return
	}
	result, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		if isNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		writeCorrelationError(c, err)
		return
	}
	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// @Summary     Delete tenant config
// @Tags        Tenant Config
// @Security    BearerAuth
// @Param       id path int true "Tenant config ID"
// @Success     204 "No content"
// @Failure     500 {object} map[string]string
// @Router      /tenant-config/{id} [delete]
func (h *TenantConfigHandler) Delete(c *gin.Context) {
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
