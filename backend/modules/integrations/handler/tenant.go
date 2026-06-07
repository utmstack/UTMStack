package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/integrations/domain"
	"github.com/utmstack/utmstack/backend/modules/integrations/dto"
	"github.com/utmstack/utmstack/backend/modules/integrations/usecase"
)

// TenantHandler exposes the per-integration tenant CRUD (panel) and the
// decrypted config feed polled by the puller plugins.
type TenantHandler struct {
	tenants *usecase.TenantUsecase
}

func NewTenantHandler(t *usecase.TenantUsecase) *TenantHandler {
	return &TenantHandler{tenants: t}
}

// @Summary     List a module's tenants (sensitive values masked)
// @Tags        Integration Tenants
// @Security    BearerAuth
// @Produce     json
// @Param       module path string true "Module name"
// @Success     200 {array}  dto.TenantResponse
// @Failure     500 {object} map[string]string
// @Router      /integrations/tenants/{module} [get]
func (h *TenantHandler) List(c *gin.Context) {
	tenants, err := h.tenants.List(c.Param("module"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tenants)
}

// @Summary     Create or update a tenant for a module
// @Tags        Integration Tenants
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       module path string           true "Module name"
// @Param       input  body dto.TenantRequest true "Tenant config"
// @Success     204 "No content"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /integrations/tenants/{module} [put]
func (h *TenantHandler) Save(c *gin.Context) {
	var req dto.TenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if err := h.tenants.Save(c.Param("module"), req); err != nil {
		// Credential-verification and validation failures are the user's to fix.
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     Delete a tenant by name
// @Tags        Integration Tenants
// @Security    BearerAuth
// @Param       module path string true "Module name"
// @Param       name   path string true "Tenant name"
// @Success     204 "No content"
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /integrations/tenants/{module}/{name} [delete]
func (h *TenantHandler) Delete(c *gin.Context) {
	err := h.tenants.Delete(c.Param("module"), c.Param("name"))
	switch {
	case errors.Is(err, domain.ErrTenantNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	default:
		c.Status(http.StatusNoContent)
	}
}
