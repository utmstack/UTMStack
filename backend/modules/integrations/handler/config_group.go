package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/utmstack/utmstack/backend/modules/integrations/domain"
	"github.com/utmstack/utmstack/backend/modules/integrations/dto"
	"github.com/utmstack/utmstack/backend/modules/integrations/usecase"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

// ConfigGroupHandler exposes one integration's configured connector instances.
// Every route serves the caller's own tenant: the integration name comes from
// the path, the tenant never does.
type ConfigGroupHandler struct {
	groups *usecase.ConfigGroupUsecase
}

func NewConfigGroupHandler(g *usecase.ConfigGroupUsecase) *ConfigGroupHandler {
	return &ConfigGroupHandler{groups: g}
}

// @Summary     List an integration's configuration groups (secrets masked)
// @Tags        Integration Configuration
// @Security    BearerAuth
// @Produce     json
// @Param       integration path string true "Integration name"
// @Success     200 {array}  dto.ConfigGroupResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /integrations/config/{integration} [get]
func (h *ConfigGroupHandler) List(c *gin.Context) {
	groups, err := h.groups.List(c.Request.Context(), c.Param("integration"))
	if err != nil {
		writeConfigGroupError(c, err)
		return
	}
	c.JSON(http.StatusOK, groups)
}

// @Summary     Create or update a configuration group
// @Tags        Integration Configuration
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       integration path string                 true "Integration name"
// @Param       input       body dto.ConfigGroupRequest true "Group configuration"
// @Success     204 "No content"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /integrations/config/{integration} [put]
func (h *ConfigGroupHandler) Save(c *gin.Context) {
	var req dto.ConfigGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if err := h.groups.Save(c.Request.Context(), c.Param("integration"), req); err != nil {
		// Credential-verification and validation failures are the user's to fix.
		writeConfigGroupError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     Delete a configuration group by name
// @Tags        Integration Configuration
// @Security    BearerAuth
// @Param       integration path string true "Integration name"
// @Param       name        path string true "Group name"
// @Success     204 "No content"
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /integrations/config/{integration}/{name} [delete]
func (h *ConfigGroupHandler) Delete(c *gin.Context) {
	if err := h.groups.Delete(c.Request.Context(), c.Param("integration"), c.Param("name")); err != nil {
		writeConfigGroupError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func writeConfigGroupError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrConfigGroupNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, tenancy.ErrNoTenant):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		// Save reports credential and validation failures here, and those are
		// the caller's to fix; the error text is what tells them which field.
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}
