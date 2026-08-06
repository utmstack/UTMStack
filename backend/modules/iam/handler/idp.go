package handler

import (
	"errors"
	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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
	case errors.Is(err, domain.ErrSSONotEntitled):
		c.JSON(http.StatusPaymentRequired, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrIDPIDForbidden), errors.Is(err, domain.ErrIDPIDRequired),
		errors.Is(err, domain.ErrIDPInvalidInput), errors.Is(err, domain.ErrIDPTypeUnsupported), errors.Is(err, domain.ErrIDPKeyRequired):
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
	audit.Record(c, audit_connectors.Event{Action: "idp.update", ResourceType: "identity_provider", ResourceID: req.ID.String()},
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
	id, err := uuid.Parse(c.Param("id"))
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
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	derr := h.uc.Delete(c.Request.Context(), id)
	audit.Record(c, audit_connectors.Event{Action: "idp.delete", ResourceType: "identity_provider", ResourceID: id.String()},
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

func (h *IdentityProviderHandler) ListMappings(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid identity provider id"})
		return
	}
	rows, err := h.uc.ListMappings(c.Request.Context(), id)
	if err != nil {
		h.writeError(c, err)
		return
	}
	if rows == nil {
		rows = []domain.IdentityProviderGroupMapping{}
	}
	c.JSON(http.StatusOK, rows)
}

// FederationHandler drives the round trip for the redirecting protocols. The
// state is kept in a short-lived cookie rather than in memory, so a callback can
// land on a different replica than the redirect left from.
type FederationHandler struct {
	uc connectors.FederationUsecase
}

func NewFederationHandler(uc connectors.FederationUsecase) *FederationHandler {
	return &FederationHandler{uc: uc}
}

const ssoStateCookie = "utm_sso_state"

// Start godoc
//
//	@Summary		Begin single sign-on
//	@Description	Redirects the browser to the provider named in the path.
//	@Tags			SSO
//	@Param			name	path	string	true	"Identity provider name"
//	@Success		302		"Redirect to the provider"
//	@Router			/sso/{name}/login [get]
func (h *FederationHandler) Start(c *gin.Context) {
	name := c.Param("name")
	redirectURL, state, err := h.uc.StartURL(c.Request.Context(), name)
	if err != nil {
		// The browser only ever sees error=sso, so the reason has to land
		// somewhere: a misconfigured issuer or an unreachable metadata URL is
		// otherwise indistinguishable from a wrong provider name.
		_ = catcher.Error("sso: cannot start sign-in", err, map[string]any{"provider": name})
		c.Redirect(http.StatusFound, appRedirect(c, "error=sso"))
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(ssoStateCookie, state, 600, "/", "", c.Request.TLS != nil, true)
	c.Redirect(http.StatusFound, redirectURL)
}

// ACS godoc
//
//	@Summary		SAML assertion consumer service
//	@Tags			SSO
//	@Accept			x-www-form-urlencoded
//	@Param			name	path	string	true	"Identity provider name"
//	@Success		302		"Redirect to the app with ?token=…"
//	@Router			/sso/{name}/acs [post]
func (h *FederationHandler) ACS(c *gin.Context) {
	name := c.Param("name")
	lc := connectors.LoginContext{IP: c.ClientIP(), UserAgent: c.Request.UserAgent()}

	want, _ := c.Cookie(ssoStateCookie)
	c.SetCookie(ssoStateCookie, "", -1, "/", "", c.Request.TLS != nil, true)

	pair, err := h.uc.ConsumeSAML(c.Request.Context(), name, c.Request, want, lc)
	audit.Record(c, audit_connectors.Event{Action: "auth.sso", ResourceType: "identity_provider", ResourceID: name},
		audit_domain.AUTH_ATTEMPT, audit_domain.AUTH_SUCCESS, err)
	if err != nil {
		_ = catcher.Error("sso: sign-in failed", err, map[string]any{"provider": name})
		c.Redirect(http.StatusFound, appRedirect(c, "error=sso"))
		return
	}
	c.Redirect(http.StatusFound, appRedirect(c, "token="+url.QueryEscape(pair.AccessToken)))
}

// Callback godoc
//
//	@Summary		OpenID Connect callback
//	@Tags			SSO
//	@Param			name	path	string	true	"Identity provider name"
//	@Param			code	query	string	true	"Authorization code"
//	@Param			state	query	string	true	"Opaque state"
//	@Success		302		"Redirect to the app with ?token=…"
//	@Router			/sso/{name}/callback [get]
func (h *FederationHandler) Callback(c *gin.Context) {
	name := c.Param("name")
	lc := connectors.LoginContext{IP: c.ClientIP(), UserAgent: c.Request.UserAgent()}

	want, _ := c.Cookie(ssoStateCookie)
	c.SetCookie(ssoStateCookie, "", -1, "/", "", c.Request.TLS != nil, true)

	pair, err := h.uc.ConsumeOIDC(c.Request.Context(), name, c.Query("code"), c.Query("state"), want, lc)
	audit.Record(c, audit_connectors.Event{Action: "auth.sso", ResourceType: "identity_provider", ResourceID: name},
		audit_domain.AUTH_ATTEMPT, audit_domain.AUTH_SUCCESS, err)
	if err != nil {
		_ = catcher.Error("sso: sign-in failed", err, map[string]any{"provider": name})
		c.Redirect(http.StatusFound, appRedirect(c, "error=sso"))
		return
	}
	c.Redirect(http.StatusFound, appRedirect(c, "token="+url.QueryEscape(pair.AccessToken)))
}

// AppBaseURL is where the browser is sent after a provider hands it back. It is
// configured rather than taken from the request because the callback arrives at
// the API, and the API is not always the same origin as the app — behind one
// nginx it is, in development and on a split deployment it is not.
var AppBaseURL string

const appLoginPath = "/auth/login"

// appRedirect builds an absolute URL back to the app root carrying the given
// query, falling back to the request's own host when nothing is configured.
func appRedirect(c *gin.Context, query string) string {
	// Back to the sign-in route, not the app root: the root is behind the route
	// guard, which redirects an unauthenticated visit and takes the token in the
	// query with it.
	if AppBaseURL != "" {
		return strings.TrimRight(AppBaseURL, "/") + appLoginPath + "?" + query
	}
	scheme := c.GetHeader("X-Forwarded-Proto")
	if scheme == "" {
		if c.Request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	return scheme + "://" + c.Request.Host + appLoginPath + "?" + query
}
