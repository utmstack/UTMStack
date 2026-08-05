package middleware

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	jwtpkg "github.com/utmstack/utmstack/backend/pkg/jwt"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

const InternalKeyHeader = "X-Internal-Key"

const TenantHeader = "X-Tenant-Id"

// Actor is re-exported for callers that already import the middleware package.
// New code should depend on pkg/authz directly.
type Actor = authz.Actor

type SupportOfFunc func(ctx context.Context, tenantID string) (string, error)

func EstablishActor(c *gin.Context, a Actor) bool { return setActor(c, a, nil) }

func setActor(c *gin.Context, a Actor, supportOf SupportOfFunc) bool {
	if !a.Internal && !resolveTenancy(c, &a, supportOf) {
		return false
	}

	c.Set("user_id", a.UserID)
	c.Set("user_login", a.Login)
	c.Set("user_email", a.Email)
	c.Set("user_permissions", a.Permissions)
	c.Set("user_roles", a.Roles)
	c.Set("session_id", a.SessionID)
	c.Set("tenant_id", a.TenantID)

	c.Set("support_access", a.Support)

	ctx := authz.WithTenantID(c.Request.Context(), a.TenantID)

	if a.Internal && a.TenantID == "" {
		ctx = tenancy.WithAllTenantsRead(ctx)
	}

	c.Request = c.Request.WithContext(ctx)

	return true
}

func resolveTenancy(c *gin.Context, a *Actor, supportOf SupportOfFunc) bool {
	if host, ok := c.Get(HostTenantKey); ok {
		if h, _ := host.(string); h != "" && a.TenantID != h {
			if !enterAsSupport(c, a, h, c.GetString(HostSupportKey)) {
				AbortUnauthorized(c, "credential does not belong to this tenant")
				return false
			}
			return true
		}
	}

	target := c.GetHeader(TenantHeader)
	if supportOf == nil || target == "" || target == a.TenantID {
		return true
	}

	level, err := supportOf(c.Request.Context(), target)
	if err != nil || !enterAsSupport(c, a, target, level) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "this tenant has not granted the platform access",
		})
		return false
	}
	return true
}

func enterAsSupport(c *gin.Context, a *Actor, tenant, level string) bool {
	if !authz.IsPlatform(a) {
		return false
	}

	if a.SessionID == 0 {
		return false
	}

	switch level {
	case authz.SupportRead:
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead &&
			c.Request.Method != http.MethodOptions {
			return false
		}
		a.Permissions = authz.ReadOnlyPermissions(a.Permissions)
	case authz.SupportFull:
	default:
		return false
	}

	a.TenantID = tenant
	a.Support = level
	return true
}

// ActorFromGin reconstructs an *authz.Actor from the values stashed in the
// gin.Context by Authenticate. Returns nil if no actor has been set.
func ActorFromGin(c *gin.Context) *authz.Actor {
	if c == nil {
		return nil
	}
	uidRaw, hasUID := c.Get("user_id")
	loginRaw, hasLogin := c.Get("user_login")
	internal := c.GetBool("internal")
	if !hasUID && !hasLogin && !internal {
		return nil
	}
	var uid uint64
	if v, ok := uidRaw.(uint64); ok {
		uid = v
	}
	login, _ := loginRaw.(string)
	email := c.GetString("user_email")
	sid := c.GetUint64("session_id")
	permsRaw, _ := c.Get("user_permissions")
	rolesRaw, _ := c.Get("user_roles")
	perms, _ := permsRaw.([]string)
	roles, _ := rolesRaw.([]string)
	tenantID := c.GetString("tenant_id")
	return &authz.Actor{
		UserID:      uid,
		Login:       login,
		Email:       email,
		Roles:       roles,
		Permissions: perms,
		SessionID:   sid,
		Internal:    internal,
		TenantID:    tenantID,
		Support:     c.GetString("support_access"),
	}
}

type APIKeyAuthFunc func(ctx context.Context, apiKey, clientIP string) (*Actor, error)

func UserAuth(signer *jwtpkg.Signer) gin.HandlerFunc {
	return Authenticate(signer, nil, "", nil)
}

func Authenticate(signer *jwtpkg.Signer, apiKeyAuth APIKeyAuthFunc, internalKey string, supportOf SupportOfFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if token, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer "); ok && token != "" {
			claims, err := signer.Verify(token)
			if err != nil {
				AbortUnauthorized(c, "invalid token")
				return
			}
			userID, err := claims.UserID()
			if err != nil {
				AbortUnauthorized(c, "invalid token subject")
				return
			}
			if !setActor(c, Actor{
				UserID:      userID,
				Login:       claims.Login,
				Email:       claims.Email,
				Roles:       claims.Roles,
				Permissions: claims.Permissions,
				SessionID:   claims.SessionID,
				TenantID:    claims.TenantID,
			}, supportOf) {
				return
			}
			c.Next()
			return
		}
		if HasInternalKey(c, internalKey) {
			c.Set("internal", true)
			if !setActor(c, Actor{
				Login:    "internal",
				Internal: true,
				TenantID: c.GetHeader(TenantHeader),
			}, supportOf) {
				return
			}
			c.Next()
			return
		}
		if apiKeyAuth != nil {
			if apiKey := c.GetHeader("Utm-Api-Key"); apiKey != "" {
				actor, err := apiKeyAuth(c.Request.Context(), apiKey, c.ClientIP())
				if err != nil || actor == nil {
					AbortUnauthorized(c, "invalid api key")
					return
				}
				if !setActor(c, *actor, supportOf) {
					return
				}
				c.Next()
				return
			}
		}
		AbortUnauthorized(c, "missing Authorization header")
	}
}

func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		actor := ActorFromGin(c)
		if err := authz.RequirePermission(actor, permission); err != nil {
			abortAuthzErr(c, err, "missing required permission: "+permission)
			return
		}
		c.Next()
	}
}

func RequireEnterprise(isEnterprise func() bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		actor := ActorFromGin(c)
		if err := authz.RequireEnterprise(actor, isEnterprise); err != nil {
			abortAuthzErr(c, err, "This feature requires an Enterprise license")
			return
		}
		c.Next()
	}
}

func RequireMSSP(isMSSP func() bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		actor := ActorFromGin(c)
		if err := authz.RequireMSSP(actor, isMSSP); err != nil {
			abortAuthzErr(c, err, "This feature requires an MSSP license")
			return
		}
		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc { return RequireRole(authz.RoleAdmin) }

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		actor := ActorFromGin(c)
		if err := authz.RequireRole(actor, role); err != nil {
			abortAuthzErr(c, err, "missing required role: "+role)
			return
		}
		c.Next()
	}
}

func RequirePlatform() gin.HandlerFunc {
	return func(c *gin.Context) {
		actor := ActorFromGin(c)
		if err := authz.RequirePlatform(actor); err != nil {
			abortAuthzErr(c, err, "this operation is limited to administrators of the default tenant")
			return
		}
		c.Next()
	}
}

func RequireOwnTenant(param string) gin.HandlerFunc {
	return func(c *gin.Context) {
		actor := ActorFromGin(c)
		if actor == nil {
			AbortUnauthorized(c, "missing authorization context")
			return
		}
		if actor.TenantID != c.Param(param) || authz.IsSupportSession(actor) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "this operation is limited to your own tenant"})
			return
		}
		c.Next()
	}
}

func RequireInternal() gin.HandlerFunc {
	return func(c *gin.Context) {
		actor := ActorFromGin(c)
		if err := authz.RequireInternal(actor); err != nil {
			abortAuthzErr(c, err, "internal-only endpoint")
			return
		}
		c.Next()
	}
}

// HasInternalKey reports whether the request carries the shared internal key.
// Authenticate and ResolveTenant both need to agree on what counts as an
// internal caller: if they diverged, a request could be internal to one and a
// tenant's to the other.
func HasInternalKey(c *gin.Context, internalKey string) bool {
	if internalKey == "" {
		return false
	}
	k := c.GetHeader(InternalKeyHeader)
	return k != "" && subtle.ConstantTimeCompare([]byte(k), []byte(internalKey)) == 1
}

func abortAuthzErr(c *gin.Context, err error, msg string) {
	if errors.Is(err, authz.ErrUnauthenticated) {
		AbortUnauthorized(c, "missing authorization context")
		return
	}
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": msg})
}

func AbortUnauthorized(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": msg})
}
