package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/integrations/connectors"
	"github.com/utmstack/utmstack/backend/modules/integrations/domain"
	"github.com/utmstack/utmstack/backend/modules/integrations/dto"
)

// ModuleHandler exposes the module catalog: list, get, categories, isActive and
// activate/deactivate. Tenant config (groups/decrypt) is served elsewhere.
type ModuleHandler struct {
	usecase connectors.ModuleUsecase
}

func NewModuleHandler(uc connectors.ModuleUsecase) *ModuleHandler {
	return &ModuleHandler{usecase: uc}
}

// @Summary     Activate or deactivate an integration module
// @Tags        Integrations
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body dto.ModuleActivationRequest true "Module name and activation status"
// @Success     200 {object} dto.ModuleResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /integrations/activate [put]
func (h *ModuleHandler) ActivateDeactivate(c *gin.Context) {
	var req dto.ModuleActivationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.usecase.ActivateDeactivate(c.Request.Context(), req)
	if err != nil {
		writeModuleError(c, "module.activateDeactivate", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     List integration modules
// @Tags        Integrations
// @Security    BearerAuth
// @Produce     json
// @Param       moduleCategory query string  false "Filter by category"
// @Param       moduleActive   query bool    false "Filter by active status"
// @Param       nameContains   query string  false "Partial match on module name"
// @Param       page           query int     false "Page (0-based)"
// @Param       size           query int     false "Page size"
// @Success     200 {array}  dto.ModuleResponse
// @Header      200 {string} X-Total-Count "Total number of records"
// @Failure     500 {object} map[string]string
// @Router      /integrations [get]
func (h *ModuleHandler) List(c *gin.Context) {
	filter := connectors.ModuleListFilter{}
	if v := c.Query("moduleCategory"); v != "" {
		filter.ModuleCategory = &v
	}
	if v := c.Query("moduleActive"); v != "" {
		b := strings.EqualFold(v, "true")
		filter.ModuleActive = &b
	}
	if v := c.Query("nameContains"); v != "" {
		filter.NameContains = &v
	}
	if v := c.Query("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			filter.Page = p
		}
	}
	if v := c.Query("size"); v != "" {
		if s, err := strconv.Atoi(v); err == nil {
			filter.Size = s
		}
	}

	items, total, err := h.usecase.List(c.Request.Context(), filter)
	if err != nil {
		writeModuleError(c, "module.list", err)
		return
	}
	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, items)
}

// @Summary     Get integration module by ID
// @Tags        Integrations
// @Security    BearerAuth
// @Produce     json
// @Param       id path int true "Module ID"
// @Success     200 {object} dto.ModuleResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /integrations/{id} [get]
func (h *ModuleHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	resp, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		writeModuleError(c, "module.get", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     List integration module categories
// @Tags        Integrations
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array}  string
// @Failure     500 {object} map[string]string
// @Router      /integrations/categories [get]
func (h *ModuleHandler) Categories(c *gin.Context) {
	resp, err := h.usecase.Categories(c.Request.Context())
	if err != nil {
		writeModuleError(c, "module.categories", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     List the data-type catalog
// @Description Each integration owns one data_type; this is the catalog the rule
// @Description editor uses to pick which data types a correlation rule targets.
// @Tags        Integrations
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array}  dto.DataTypeOption
// @Failure     500 {object} map[string]string
// @Router      /integrations/data-types [get]
func (h *ModuleHandler) DataTypes(c *gin.Context) {
	resp, err := h.usecase.DataTypes(c.Request.Context())
	if err != nil {
		writeModuleError(c, "module.dataTypes", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Report whether an integration module is active
// @Tags        Integrations
// @Security    BearerAuth
// @Produce     json
// @Param       moduleName query string true "Module name"
// @Success     200 {object} map[string]bool
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /integrations/is-active [get]
func (h *ModuleHandler) IsActive(c *gin.Context) {
	moduleName := c.Query("moduleName")
	if moduleName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing moduleName"})
		return
	}
	resp, err := h.usecase.IsActive(c.Request.Context(), moduleName)
	if err != nil {
		writeModuleError(c, "module.isActive", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// writeModuleError maps module-catalog domain sentinels to HTTP statuses.
func writeModuleError(c *gin.Context, op string, err error) {
	switch {
	case errors.Is(err, domain.ErrModuleNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrInvalidActivation):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrSystemModule):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		_ = catcher.Error(op, err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "operation failed"})
	}
}

// @Summary     Create a custom integration
// @Tags        Integrations
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.CreateModuleRequest true "Custom integration"
// @Success     201 {object} dto.ModuleResponse
// @Failure     400 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /integrations [post]
func (h *ModuleHandler) Create(c *gin.Context) {
	var req dto.CreateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	resp, err := h.usecase.Create(c.Request.Context(), req)
	if err != nil {
		writeModuleError(c, "module.create", err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// @Summary     Update a custom integration's metadata
// @Tags        Integrations
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id    path int                    true "Module ID"
// @Param       input body dto.UpdateModuleRequest true "Metadata"
// @Success     200 {object} dto.ModuleResponse
// @Failure     400 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /integrations/{id} [put]
func (h *ModuleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req dto.UpdateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	resp, err := h.usecase.Update(c.Request.Context(), id, req)
	if err != nil {
		writeModuleError(c, "module.update", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Delete a custom integration
// @Tags        Integrations
// @Security    BearerAuth
// @Param       id path int true "Module ID"
// @Success     204 "No content"
// @Failure     400 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /integrations/{id} [delete]
func (h *ModuleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		writeModuleError(c, "module.delete", err)
		return
	}
	c.Status(http.StatusNoContent)
}
