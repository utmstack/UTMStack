package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
)

// SelfHealDefaultDomain fills the default tenant's Domain from the first
// request that reaches the API when it was left blank at install time.
// The heal callback owns fast-path skipping — this middleware just hands it
// the best hostname it can see (proxy-forwarded first, direct Host next).
func SelfHealDefaultDomain(heal func(ctx context.Context, host string) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		if heal != nil {
			host := c.Request.Header.Get("X-Forwarded-Host")
			if host == "" {
				host = c.Request.Host
			}
			_ = heal(c.Request.Context(), host)
		}
		c.Next()
	}
}
