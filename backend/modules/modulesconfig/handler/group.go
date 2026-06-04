package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/dto"
)

type GroupHandler struct {
	usecase connectors.GroupUsecase
}

func NewGroupHandler(uc connectors.GroupUsecase) *GroupHandler {
	return &GroupHandler{usecase: uc}
}

// @Summary Create a configuration group
// @Tags ModulesConfig
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body dto.CreateModuleGroupRequest true "Group definition"
// @Success 200 {object} dto.ModuleGroupResponse
// @Router /utm-configuration-groups [post]
func (h *GroupHandler) Create(c *gin.Context) {
	var req dto.CreateModuleGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.usecase.Create(c.Request.Context(), req)
	if err != nil {
		writeError(c, "group.create", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary Update a configuration group
// @Tags ModulesConfig
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body dto.UpdateModuleGroupRequest true "Group update"
// @Success 200 {object} dto.ModuleGroupResponse
// @Router /utm-configuration-groups [put]
func (h *GroupHandler) Update(c *gin.Context) {
	var req dto.UpdateModuleGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.usecase.Update(c.Request.Context(), req)
	if err != nil {
		writeError(c, "group.update", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *GroupHandler) ListByModuleID(c *gin.Context) {
	moduleID, err := strconv.ParseInt(c.Query("moduleId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid moduleId"})
		return
	}
	resp, err := h.usecase.ListByModuleID(c.Request.Context(), moduleID)
	if err != nil {
		writeError(c, "group.list", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *GroupHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	resp, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		writeError(c, "group.get", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *GroupHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		writeError(c, "group.delete", err)
		return
	}
	c.Status(http.StatusNoContent)
}
