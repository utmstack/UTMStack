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

	own := api.Group("/tenants", userAuth, mssp)
	ownTenant := []gin.HandlerFunc{middleware.RequireAdmin(), middleware.RequireOwnTenant("id")}
	own.GET("/:id/support-access", append(ownTenant, h.GetSupportAccess)...)
	own.PUT("/:id/support-access", append(ownTenant, h.SetSupportAccess)...)
}
