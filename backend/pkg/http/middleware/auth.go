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

func RequireEnterprise(isEnterprise func() bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetBool("internal") {
			c.Next()
			return
		}
		if isEnterprise() {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "This feature requires an Enterprise license"})
	}
}

func RequireMSSP(isMSSP func() bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetBool("internal") {
			c.Next()
			return
		}
		if isMSSP() {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "This feature requires an MSSP license"})
	}
}

func RequireAdmin() gin.HandlerFunc {
	return RequireRole("ROLE_ADMIN")
}

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

func RequireInternal() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetBool("internal") {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "internal-only endpoint"})
	}
}

func AbortUnauthorized(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": msg})
}
