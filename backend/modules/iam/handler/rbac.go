package handler

import (
	"errors"
	"github.com/google/uuid"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
)

type RoleHandler struct {
	roleUsecase connectors.RoleUsecase
}

func NewRoleHandler(roleUsecase connectors.RoleUsecase) *RoleHandler {
	return &RoleHandler{roleUsecase: roleUsecase}
}

// @Summary  List roles
// @Tags     Roles
// @Security BearerAuth
// @Produce  json
// @Success  200 {array} dto.RoleResponse
// @Failure  401 {object} map[string]string
// @Failure  403 {object} map[string]string
// @Failure  500 {object} map[string]string
// @Router   /roles [get]
func (h *RoleHandler) List(c *gin.Context) {
	resp, err := h.roleUsecase.List(c.Request.Context())
	if err != nil {
		_ = catcher.Error("list roles failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list roles"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary  Get role
// @Tags     Roles
// @Security BearerAuth
// @Produce  json
// @Param    name path string true "Role name (e.g. ROLE_ADMIN)"
// @Success  200 {object} dto.RoleDetailResponse
// @Failure  401 {object} map[string]string
// @Failure  403 {object} map[string]string
// @Failure  404 {object} map[string]string
// @Router   /roles/{name} [get]
func (h *RoleHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}
	resp, err := h.roleUsecase.Get(c.Request.Context(), id)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req dto.RoleUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.roleUsecase.Create(c.Request.Context(), req)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *RoleHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}
	var req dto.RoleUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.roleUsecase.Update(c.Request.Context(), id, req)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}
	if err := h.roleUsecase.Delete(c.Request.Context(), id); err != nil {
		h.writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *RoleHandler) ListPermissions(c *gin.Context) {
	resp, err := h.roleUsecase.ListPermissions(c.Request.Context())
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *RoleHandler) writeErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrRoleNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrRoleImmutable):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrRoleNameTaken), errors.Is(err, domain.ErrPermissionNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}

func writeRoleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrRoleNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
	default:
		_ = catcher.Error("role op failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "operation failed"})
	}
}
