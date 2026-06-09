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
)

const InternalKeyHeader = "X-Internal-Key"

// Actor is re-exported for callers that already import the middleware package.
// New code should depend on pkg/authz directly.
type Actor = authz.Actor

func setActor(c *gin.Context, a Actor) {
	c.Set("user_id", a.UserID)
	c.Set("user_login", a.Login)
	c.Set("user_email", a.Email)
	c.Set("user_permissions", a.Permissions)
	c.Set("user_roles", a.Roles)
	c.Set("session_id", a.SessionID)
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
	return &authz.Actor{
		UserID:      uid,
		Login:       login,
		Email:       email,
		Roles:       roles,
		Permissions: perms,
		SessionID:   sid,
		Internal:    internal,
	}
}

type APIKeyAuthFunc func(ctx context.Context, apiKey, clientIP string) (*Actor, error)

func UserAuth(signer *jwtpkg.Signer) gin.HandlerFunc {
	return Authenticate(signer, nil, "")
}

func Authenticate(signer *jwtpkg.Signer, apiKeyAuth APIKeyAuthFunc, internalKey string) gin.HandlerFunc {
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
			setActor(c, Actor{
				UserID:      userID,
				Login:       claims.Login,
				Email:       claims.Email,
				Roles:       claims.Roles,
				Permissions: claims.Permissions,
				SessionID:   claims.SessionID,
			})
			c.Next()
			return
		}
		if internalKey != "" {
			if k := c.GetHeader(InternalKeyHeader); k != "" &&
				subtle.ConstantTimeCompare([]byte(k), []byte(internalKey)) == 1 {
				c.Set("internal", true)
				setActor(c, Actor{Login: "internal", Internal: true})
				c.Next()
				return
			}
		}
		if apiKeyAuth != nil {
			if apiKey := c.GetHeader("Utm-Api-Key"); apiKey != "" {
				actor, err := apiKeyAuth(c.Request.Context(), apiKey, c.ClientIP())
				if err != nil || actor == nil {
					AbortUnauthorized(c, "invalid api key")
					return
				}
				setActor(c, *actor)
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
