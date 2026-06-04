package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/dto"
)

type ModuleHandler struct {
	usecase connectors.ModuleUsecase
}

func NewModuleHandler(uc connectors.ModuleUsecase) *ModuleHandler {
	return &ModuleHandler{usecase: uc}
}

// @Summary Activate or deactivate a module
// @Tags ModulesConfig
// @Security BearerAuth
// @Produce json
// @Param serverId query int true "Server id"
// @Param nameShort query string true "Module name (kind)"
// @Param activationStatus query bool true "Target activation status"
// @Success 200 {object} dto.ModuleResponse
// @Router /utm-modules/activateDeactivate [put]
func (h *ModuleHandler) ActivateDeactivate(c *gin.Context) {
	var req dto.ModuleActivationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.usecase.ActivateDeactivate(c.Request.Context(), req)
	if err != nil {
		writeError(c, "module.activateDeactivate", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary List modules
// @Tags ModulesConfig
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.ModuleResponse
// @Router /utm-modules [get]
func (h *ModuleHandler) List(c *gin.Context) {
	filter := connectors.ModuleListFilter{}
	if v := c.Query("serverId"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			filter.ServerID = &id
		}
	}
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
		writeError(c, "module.list", err)
		return
	}
	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, items)
}

// @Summary Get module by id
// @Tags ModulesConfig
// @Security BearerAuth
// @Produce json
// @Param id path int true "Module id"
// @Success 200 {object} dto.ModuleResponse
// @Router /utm-modules/{id} [get]
func (h *ModuleHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	resp, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		writeError(c, "module.get", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ModuleDetails returns the module's persisted state. Sensitive config values
// are masked.
func (h *ModuleHandler) ModuleDetails(c *gin.Context) {
	serverID, err := strconv.ParseInt(c.Query("serverId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid serverId"})
		return
	}
	nameShort := c.Query("nameShort")
	if nameShort == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing nameShort"})
		return
	}
	resp, err := h.usecase.Details(c.Request.Context(), serverID, nameShort, false)
	if err != nil {
		writeError(c, "module.details", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ModuleDetailsDecrypted is the internal-key-gated endpoint event-processor and
// other internal services use to read raw configuration values. The route
// configuration on this handler is what enforces the gate — the handler itself
// trusts that the middleware admitted only internal callers.
func (h *ModuleHandler) ModuleDetailsDecrypted(c *gin.Context) {
	if !c.GetBool("internal") {
		c.JSON(http.StatusForbidden, gin.H{"error": "internal key required"})
		return
	}
	nameShort := c.Query("nameShort")
	if nameShort == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing nameShort"})
		return
	}
	serverID := int64(0)
	if v := c.Query("serverId"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			serverID = id
		}
	}
	resp, err := h.usecase.Details(c.Request.Context(), serverID, nameShort, true)
	if err != nil {
		writeError(c, "module.detailsDecrypted", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CheckRequirements delegates to the ModuleKind's preflight checks. Returns
// 422 if the kind is not registered (i.e. not yet ported).
func (h *ModuleHandler) CheckRequirements(c *gin.Context) {
	serverID, err := strconv.ParseInt(c.Query("serverId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid serverId"})
		return
	}
	nameShort := c.Query("nameShort")
	if nameShort == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing nameShort"})
		return
	}
	resp, err := h.usecase.CheckRequirements(c.Request.Context(), serverID, nameShort)
	if err != nil {
		writeError(c, "module.checkRequirements", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ModuleHandler) Categories(c *gin.Context) {
	var serverID *int64
	if v := c.Query("serverId"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			serverID = &id
		}
	}
	resp, err := h.usecase.Categories(c.Request.Context(), serverID)
	if err != nil {
		writeError(c, "module.categories", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ModuleHandler) IsActive(c *gin.Context) {
	moduleName := c.Query("moduleName")
	if moduleName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing moduleName"})
		return
	}
	resp, err := h.usecase.IsActive(c.Request.Context(), moduleName)
	if err != nil {
		writeError(c, "module.isActive", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
