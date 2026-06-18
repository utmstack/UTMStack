package handler

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
)

type SAMLHandler struct {
	uc connectors.SAMLUsecase
}

func NewSAMLHandler(uc connectors.SAMLUsecase) *SAMLHandler {
	return &SAMLHandler{uc: uc}
}

// Initiate godoc
//
//	@Summary		Start SAML SSO login
//	@Description	Redirects the browser to the IdP with a signed AuthnRequest for the named provider.
//	@Tags			SSO
//	@Param			name	path	string	true	"Identity provider name"
//	@Success		302		"Redirect to IdP"
//	@Failure		302		"Redirect to the app with ?error=saml2"
//	@Router			/sso/saml/{name}/login [get]
func (h *SAMLHandler) Initiate(c *gin.Context) {
	name := c.Param("name")
	redirectURL, err := h.uc.InitiateURL(c.Request.Context(), name)
	if err != nil {
		c.Redirect(http.StatusFound, appRedirect(c, "error=saml2"))
		return
	}
	c.Redirect(http.StatusFound, redirectURL)
}

// ACS godoc
//
//	@Summary		SAML assertion consumer service
//	@Description	Receives the IdP SAMLResponse, validates it, maps the NameID to a local user and redirects to the app with a session token.
//	@Tags			SSO
//	@Accept			x-www-form-urlencoded
//	@Param			name	path		string	true	"Identity provider name"
//	@Param			SAMLResponse	formData	string	true	"Base64 SAML response"
//	@Success		302		"Redirect to the app with ?token=…"
//	@Failure		302		"Redirect to the app with ?error=saml2"
//	@Router			/sso/saml/{name}/acs [post]
func (h *SAMLHandler) ACS(c *gin.Context) {
	name := c.Param("name")
	lc := connectors.LoginContext{IP: c.ClientIP(), UserAgent: c.Request.UserAgent()}

	token, err := h.uc.ConsumeACS(c.Request.Context(), name, c.Request, lc)
	audit.Record(c, audit_connectors.Event{Action: "auth.saml", ResourceType: "identity_provider", ResourceID: name},
		audit_domain.AUTH_ATTEMPT, audit_domain.AUTH_SUCCESS, err)
	if err != nil {
		c.Redirect(http.StatusFound, appRedirect(c, "error=saml2"))
		return
	}
	c.Redirect(http.StatusFound, appRedirect(c, "token="+url.QueryEscape(token)))
}

// appRedirect builds an absolute URL back to the app root (same host as the
// request, honoring the proxy's X-Forwarded-Proto) carrying the given query.
func appRedirect(c *gin.Context, query string) string {
	scheme := c.GetHeader("X-Forwarded-Proto")
	if scheme == "" {
		if c.Request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	return scheme + "://" + c.Request.Host + "/?" + query
}
