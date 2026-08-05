package tenant

import (
	"github.com/gin-gonic/gin"

	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth, mssp, platform gin.HandlerFunc) {
	h := m.GetTenantHandler()

	read := middleware.RequirePermission("tenant.read")
	write := middleware.RequirePermission("tenant.write")

	g := api.Group("/tenants", userAuth, mssp, platform)

	g.GET("", read, h.List)
	g.GET("/:id", read, h.GetByID)
	g.POST("", write, h.Create)
	g.PUT("/:id", write, h.Update)
	g.DELETE("/:id", write, h.Terminate)

	// Outside the platform gate: the one thing about a tenant the platform may
	// not decide.
	own := api.Group("/tenants", userAuth, mssp)
	own.PUT("/:id/support-access", middleware.RequireAdmin(), middleware.RequireOwnTenant("id"), h.SetSupportAccess)
}
