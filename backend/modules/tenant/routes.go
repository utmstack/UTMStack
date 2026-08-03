package tenant

import (
	"github.com/gin-gonic/gin"

	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

// RegisterRoutes mounts the tenant surface.
//
// The whole group sits behind the MSSP licence: on a single-tenant install the
// scoping callbacks are inert, so a tenant row would isolate nothing and the
// endpoint would only mislead whoever called it.
//
// Provisioning is internal on purpose: a tenant is created and terminated by
// whatever sells the subscription, never from inside an instance.
func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth, mssp gin.HandlerFunc) {
	h := m.GetTenantHandler()

	read := middleware.RequirePermission("tenant.read")
	write := middleware.RequirePermission("tenant.write")
	internal := middleware.RequireInternal()

	g := api.Group("/tenants", userAuth, mssp)

	g.GET("", read, h.List)
	g.GET("/:id", read, h.GetByID)
	g.PUT("/:id", write, h.Update)

	g.POST("", internal, h.Create)
	g.DELETE("/:id", internal, h.Terminate)
}
