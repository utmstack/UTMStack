package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

const HostTenantKey = "host_tenant_id"

const HostSupportKey = "host_support_access"

func ResolveTenant(isMSSP func() bool, internalKey string, resolve func(ctx context.Context, host string) (tenantID, support string, err error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		if isMSSP == nil || !isMSSP() || resolve == nil {
			c.Next()
			return
		}

		if HasInternalKey(c, internalKey) {
			c.Next()
			return
		}

		tenantID, support, err := resolve(c.Request.Context(), c.Request.Host)
		if err != nil || tenantID == "" {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "unknown tenant"})
			return
		}

		c.Set(HostTenantKey, tenantID)
		c.Set(HostSupportKey, support)
		c.Request = c.Request.WithContext(authz.WithTenantID(c.Request.Context(), tenantID))
		c.Next()
	}
}
