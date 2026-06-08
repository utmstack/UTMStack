package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
)

type IdentityProviderHandler struct {
	uc connectors.IdentityProviderUsecase
}

func NewIdentityProviderHandler(uc connectors.IdentityProviderUsecase) *IdentityProviderHandler {
	return &IdentityProviderHandler{uc: uc}
}

func (h *IdentityProviderHandler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrIDPNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrIDPIDForbidden), errors.Is(err, domain.ErrIDPIDRequired),
		errors.Is(err, domain.ErrIDPInvalidInput), errors.Is(err, domain.ErrIDPKeyRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// Create godoc
//
//	@Summary		Create a SAML identity provider
//	@Description	Registers a SAML2 IdP. The SP private key (PEM) is encrypted at rest and never returned.
//	@Tags			Identity Providers
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		dto.IdentityProviderRequest	true	"IdP config"
//	@Success		201		{object}	domain.IdentityProviderConfig
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/identity-providers [post]
func (h *IdentityProviderHandler) Create(c *gin.Context) {
	var req dto.IdentityProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.uc.Create(c.Request.Context(), req)
	audit.Record(c, audit_connectors.Event{Action: "idp.create", ResourceType: "identity_provider", ResourceID: req.Name},
		audit_domain.IDP_CONFIG_CREATE_ATTEMPT, audit_domain.IDP_CONFIG_CREATE_SUCCESS, err)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, res)
}

// Update godoc
//
//	@Summary		Update a SAML identity provider
//	@Description	Updates an IdP. Leave the private key empty to keep the stored one.
//	@Tags			Identity Providers
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		dto.IdentityProviderRequest	true	"IdP config"
//	@Success		200		{object}	domain.IdentityProviderConfig
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/identity-providers [put]
func (h *IdentityProviderHandler) Update(c *gin.Context) {
	var req dto.IdentityProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.uc.Update(c.Request.Context(), req)
	audit.Record(c, audit_connectors.Event{Action: "idp.update", ResourceType: "identity_provider", ResourceID: strconv.FormatUint(req.ID, 10)},
		audit_domain.IDP_CONFIG_UPDATE_ATTEMPT, audit_domain.IDP_CONFIG_UPDATE_SUCCESS, err)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// List godoc
//
//	@Summary		List SAML identity providers
//	@Tags			Identity Providers
//	@Security		BearerAuth
//	@Produce		json
//	@Param			name			query		string	false	"Filter by name (substring)"
//	@Param			providerType	query		string	false	"Filter by provider type"
//	@Param			active			query		bool	false	"Filter by active"
//	@Param			page			query		int		false	"Page (0-based)"
//	@Param			size			query		int		false	"Page size"
//	@Success		200				{array}		domain.IdentityProviderConfig
//	@Header			200				{string}	X-Total-Count	"Total records"
//	@Failure		500				{object}	map[string]string
//	@Router			/identity-providers [get]
func (h *IdentityProviderHandler) List(c *gin.Context) {
	var f dto.IdentityProviderFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	items, total, err := h.uc.List(c.Request.Context(), f)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	if items == nil {
		items = []domain.IdentityProviderConfig{}
	}
	c.JSON(http.StatusOK, items)
}

// GetByID godoc
//
//	@Summary		Get a SAML identity provider by id
//	@Tags			Identity Providers
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"IdP id"
//	@Success		200	{object}	domain.IdentityProviderConfig
//	@Failure		404	{object}	map[string]string
//	@Router			/identity-providers/{id} [get]
func (h *IdentityProviderHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	res, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// Delete godoc
//
//	@Summary		Delete a SAML identity provider by id
//	@Tags			Identity Providers
//	@Security		BearerAuth
//	@Param			id	path	int	true	"IdP id"
//	@Success		200	"Deleted"
//	@Failure		500	{object}	map[string]string
//	@Router			/identity-providers/{id} [delete]
func (h *IdentityProviderHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	derr := h.uc.Delete(c.Request.Context(), id)
	audit.Record(c, audit_connectors.Event{Action: "idp.delete", ResourceType: "identity_provider", ResourceID: strconv.FormatUint(id, 10)},
		audit_domain.IDP_CONFIG_DELETE_ATTEMPT, audit_domain.IDP_CONFIG_DELETE_SUCCESS, derr)
	if derr != nil {
		h.writeError(c, derr)
		return
	}
	c.Status(http.StatusOK)
}

// PublicList godoc
//
//	@Summary		List active identity providers (public)
//	@Description	Unauthenticated list of active IdPs for the login page ("Login with …" buttons).
//	@Tags			Identity Providers
//	@Produce		json
//	@Success		200	{array}		dto.IdentityProviderPublic
//	@Failure		500	{object}	map[string]string
//	@Router			/idp-providers [get]
func (h *IdentityProviderHandler) PublicList(c *gin.Context) {
	items, err := h.uc.ListActive(c.Request.Context())
	if err != nil {
		h.writeError(c, err)
		return
	}
	if items == nil {
		items = []dto.IdentityProviderPublic{}
	}
	c.JSON(http.StatusOK, items)
}
