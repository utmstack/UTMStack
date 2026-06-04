package middleware

import (
	"context"
	"crypto/subtle"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	jwtpkg "github.com/utmstack/utmstack/backend/pkg/jwt"
)

// InternalKeyHeader carries the shared INTERNAL_KEY for service-to-service calls
// (plugins, agent-manager). Matches the header the SOC-AI / internal clients send.
const InternalKeyHeader = "X-Internal-Key"

type Actor struct {
	UserID      uint64
	Login       string
	Email       string
	Roles       []string
	Permissions []string
	SessionID   uint64
}

func setActor(c *gin.Context, a Actor) {
	c.Set("user_id", a.UserID)
	c.Set("user_login", a.Login)
	c.Set("user_email", a.Email)
	c.Set("user_permissions", a.Permissions)
	c.Set("user_roles", a.Roles)
	c.Set("session_id", a.SessionID)
}

// APIKeyAuthFunc validates an api key against a client IP and returns the
// resolved actor (the key owner's identity + roles/permissions).
type APIKeyAuthFunc func(ctx context.Context, apiKey, clientIP string) (*Actor, error)

func UserAuth(signer *jwtpkg.Signer) gin.HandlerFunc {
	return Authenticate(signer, nil, "")
}

// Authenticate authenticates a request by, in order: a JWT (Authorization:
// Bearer), the shared internal key (X-Internal-Key, for service-to-service calls
// from plugins/agent-manager), or — when apiKeyAuth is set — the Utm-Api-Key
// header. internalKey is the configured INTERNAL_KEY; pass "" to disable it.
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
				// Trusted internal service (plugin / agent-manager). It carries no
				// user identity, and bypasses permission/role gates.
				c.Set("internal", true)
				setActor(c, Actor{Login: "internal"})
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
		if c.GetBool("internal") {
			c.Next()
			return
		}
		raw, ok := c.Get("user_permissions")
		if !ok {
			AbortUnauthorized(c, "missing permissions context")
			return
		}
		perms, ok := raw.([]string)
		if !ok {
			AbortUnauthorized(c, "invalid permissions context")
			return
		}
		if slices.Contains(perms, permission) {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "missing required permission: " + permission})
	}
}

// RequireAdmin gates a route to users whose JWT carries the `admin` role.
// Used for endpoints we want admin-only without adding a fine-grained
// permission to the catalog.
func RequireAdmin() gin.HandlerFunc {
	return RequireRole("ROLE_ADMIN")
}

// RequireRole accepts a role name and aborts with 403 if the caller's JWT
// does not list it. Roles are populated by UserAuth from JWT claims.
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetBool("internal") {
			c.Next()
			return
		}
		raw, ok := c.Get("user_roles")
		if !ok {
			AbortUnauthorized(c, "missing roles context")
			return
		}
		roles, ok := raw.([]string)
		if !ok {
			AbortUnauthorized(c, "invalid roles context")
			return
		}
		if slices.Contains(roles, role) {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "missing required role: " + role})
	}
}

func AbortUnauthorized(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": msg})
}
