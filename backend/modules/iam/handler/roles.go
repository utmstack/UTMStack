package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/pkg/logger"
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
		logger.Error("list roles failed: " + err.Error())
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
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing role name"})
		return
	}
	resp, err := h.roleUsecase.Get(c.Request.Context(), name)
	if err != nil {
		writeRoleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func writeRoleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrRoleNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
	default:
		logger.Error("role op failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "operation failed"})
	}
}
