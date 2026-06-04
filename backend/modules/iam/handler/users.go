package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
)

type UserHandler struct {
	userUsecase connectors.UserUsecase
}

func NewUserHandler(userUsecase connectors.UserUsecase) *UserHandler {
	return &UserHandler{userUsecase: userUsecase}
}

// @Summary     List users
// @Tags        Users
// @Security    BearerAuth
// @Produce     json
// @Param       page      query int    false "Page (default 1)"
// @Param       page_size query int    false "Page size (default 20, max 200)"
// @Param       search    query string false "Match against login/email/first_name/last_name"
// @Success     200 {object} dto.UserListResponse
// @Router      /users [get]
func (h *UserHandler) List(c *gin.Context) {
	var q dto.ListUsersQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.userUsecase.List(c.Request.Context(), q)
	if err != nil {
		_ = catcher.Error("list users failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list users"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Get user
// @Tags        Users
// @Security    BearerAuth
// @Produce     json
// @Param       id path int true "User ID"
// @Success     200 {object} dto.UserDetailResponse
// @Failure     404 {object} map[string]string
// @Router      /users/{id} [get]
func (h *UserHandler) Get(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	resp, err := h.userUsecase.Get(c.Request.Context(), id)
	if err != nil {
		writeUserError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Create user
// @Tags        Users
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.CreateUserRequest true "User"
// @Success     201 {object} dto.UserDetailResponse
// @Router      /users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var input dto.CreateUserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.userUsecase.Create(c.Request.Context(), userLoginFromCtx(c), input)
	audit.Record(c, audit_connectors.Event{Action: "user.create"}, audit_domain.USER_CREATION_ATTEMPT, audit_domain.USER_CREATION_SUCCESS, err)
	if err != nil {
		writeUserError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// @Summary     Update user
// @Tags        Users
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id    path int                  true "User ID"
// @Param       input body dto.UpdateUserRequest true "Fields to update"
// @Success     200 {object} dto.UserDetailResponse
// @Router      /users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var input dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.userUsecase.Update(c.Request.Context(), userLoginFromCtx(c), id, input)
	audit.Record(c, audit_connectors.Event{Action: "user.update"}, audit_domain.USER_UPDATE_ATTEMPT, audit_domain.USER_UPDATE_SUCCESS, err)
	if err != nil {
		writeUserError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Deactivate user
// @Tags        Users
// @Security    BearerAuth
// @Param       id path int true "User ID"
// @Success     204 "Deactivated"
// @Router      /users/{id} [delete]
func (h *UserHandler) Deactivate(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	err := h.userUsecase.Deactivate(c.Request.Context(), id)
	audit.Record(c, audit_connectors.Event{Action: "user.deactivate"}, audit_domain.USER_DELETE_ATTEMPT, audit_domain.USER_DELETE_SUCCESS, err)
	if err != nil {
		writeUserError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     Assign roles to user
// @Tags        Users
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id    path int                    true "User ID"
// @Param       input body dto.AssignRolesRequest true "Role IDs"
// @Success     204 "Roles assigned"
// @Router      /users/{id}/roles [put]
func (h *UserHandler) AssignRoles(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var input dto.AssignRolesRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.userUsecase.AssignRoles(c.Request.Context(), userLoginFromCtx(c), id, input)
	audit.Record(c, audit_connectors.Event{Action: "user.roles_assign"}, audit_domain.USER_UPDATE_ATTEMPT, audit_domain.USER_UPDATE_SUCCESS, err)
	if err != nil {
		writeUserError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func writeUserError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
	case errors.Is(err, domain.ErrLoginOrEmailTaken):
		c.JSON(http.StatusConflict, gin.H{"error": "login or email already in use"})
	case errors.Is(err, domain.ErrInvalidRoleSubset):
		c.JSON(http.StatusBadRequest, gin.H{"error": "one or more role names do not exist"})
	default:
		_ = catcher.Error("user op failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "operation failed"})
	}
}

func pathID(c *gin.Context, name string) (uint64, bool) {
	idStr := c.Param(name)
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + name})
		return 0, false
	}
	return id, true
}
