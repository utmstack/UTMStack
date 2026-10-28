package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/adaudit/connectors"
	"github.com/utmstack/utmstack/backend/modules/adaudit/dto"
	"github.com/utmstack/utmstack/backend/modules/adaudit/repository"
	"github.com/utmstack/utmstack/backend/modules/adaudit/usecase"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
	"gorm.io/gorm"
)

type IdentityHandler struct {
	uc connectors.IdentityUsecase
}

func NewIdentityHandler(uc connectors.IdentityUsecase) *IdentityHandler {
	return &IdentityHandler{uc: uc}
}

func actorIdentity(c *gin.Context) string {
	if a := middleware.ActorFromGin(c); a != nil && a.Email != "" {
		return a.Email
	}
	return "internal"
}

var ErrIdentityAssignConflict = errors.New("user is already assigned to another identity")

func writeIdentityError(c *gin.Context, op string, err error) {
	switch {
	case errors.Is(err, usecase.ErrIdentityNameInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, repository.ErrIdentityNotFound),
		errors.Is(err, repository.ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, repository.ErrIdentityNameTaken),
		errors.Is(err, repository.ErrIdentityInUse),
		errors.Is(err, repository.ErrTenantMismatch),
		errors.Is(err, ErrIdentityAssignConflict),
		errors.Is(err, gorm.ErrDuplicatedKey):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		_ = catcher.Error("adaudit: "+op+" failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not " + op + " identity"})
	}
}

// Create godoc
//
//	@Summary		Create an identity alias
//	@Tags			AD Audit
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		dto.CreateIdentityRequest	true	"Alias name"
//	@Success		200		{object}	dto.IdentityResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		409		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/ad-audit/identities [post]
func (h *IdentityHandler) Create(c *gin.Context) {
	var req dto.CreateIdentityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.uc.Create(c.Request.Context(), req.Name, actorIdentity(c))
	if err != nil {
		writeIdentityError(c, "create", err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// List godoc
//
//	@Summary		List identity aliases
//	@Tags			AD Audit
//	@Security		BearerAuth
//	@Produce		json
//	@Param			search	query		string	false	"Case-insensitive substring on name"
//	@Success		200		{array}		dto.IdentityListItem
//	@Failure		500		{object}	map[string]string
//	@Router			/ad-audit/identities [get]
func (h *IdentityHandler) List(c *gin.Context) {
	items, err := h.uc.List(c.Request.Context(), c.Query("search"))
	if err != nil {
		_ = catcher.Error("adaudit: identity list failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list identities"})
		return
	}
	c.JSON(http.StatusOK, items)
}

// Detail godoc
//
//	@Summary		Show an identity alias and its users
//	@Tags			AD Audit
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		string	true	"Identity id"
//	@Success		200	{object}	dto.IdentityDetail
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/ad-audit/identities/{id} [get]
func (h *IdentityHandler) Detail(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is not a valid id"})
		return
	}
	res, err := h.uc.Detail(c.Request.Context(), id)
	if err != nil {
		writeIdentityError(c, "load", err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// Delete godoc
//
//	@Summary		Delete an identity alias
//	@Tags			AD Audit
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		string	true	"Identity id"
//	@Success		200	{object}	map[string]int
//	@Failure		404	{object}	map[string]string
//	@Failure		409	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/ad-audit/identities/{id} [delete]
func (h *IdentityHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is not a valid id"})
		return
	}
	n, err := h.uc.Delete(c.Request.Context(), id)
	if err != nil {
		writeIdentityError(c, "delete", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": n})
}

// Assign godoc
//
//	@Summary		Assign a user to an identity alias
//	@Description	An empty or null identityId unassigns the user.
//	@Tags			AD Audit
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string			true	"User id"
//	@Param			input	body		dto.AssignIdentityRequest	true	"Identity to assign"
//	@Success		200		{object}	map[string]bool
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		409		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/ad-audit/users/{id}/identity [put]
func (h *IdentityHandler) Assign(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is not a valid id"})
		return
	}
	var req dto.AssignIdentityRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "body must be a JSON object with an 'identityId' string"})
			return
		}
	}
	if req.IdentityID == "" {
		if err := h.uc.Unassign(c.Request.Context(), userID, actorIdentity(c)); err != nil {
			writeIdentityError(c, "assign", err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"assigned": false})
		return
	}
	aliasID, err := uuid.Parse(req.IdentityID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "identityId is not a valid id"})
		return
	}
	if err := h.uc.Assign(c.Request.Context(), userID, aliasID, actorIdentity(c)); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			c.JSON(http.StatusConflict, gin.H{"error": ErrIdentityAssignConflict.Error()})
			return
		}
		writeIdentityError(c, "assign", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"assigned": true})
}

// Unassign godoc
//
//	@Summary		Remove a user from its identity alias
//	@Tags			AD Audit
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		string	true	"User id"
//	@Success		200	{object}	map[string]bool
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/ad-audit/users/{id}/identity [patch]
func (h *IdentityHandler) Unassign(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is not a valid id"})
		return
	}
	if err := h.uc.Unassign(c.Request.Context(), userID, actorIdentity(c)); err != nil {
		writeIdentityError(c, "assign", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"assigned": false})
}
