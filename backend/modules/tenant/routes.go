package tenant

import (
	"github.com/gin-gonic/gin"

	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth, mssp, platform gin.HandlerFunc) {
	h := m.GetTenantHandler()

	read := middleware.RequirePermission("tenant.read")
	write := middleware.RequirePermission("tenant.write")
	internal := middleware.RequireInternal()

	g := api.Group("/tenants", userAuth, mssp, platform)

	g.GET("", read, h.List)
	g.GET("/:id", read, h.GetByID)
	g.PUT("/:id", write, h.Update)

	g.POST("", internal, h.Create)
	g.DELETE("/:id", internal, h.Terminate)
}
