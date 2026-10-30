package adaudit

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	h := m.Handler()
	g := api.Group("/ad-audit", userAuth)

	g.POST("/users", middleware.RequireInternal(), h.Ingest)
	g.GET("/users/sync", middleware.RequireInternal(), h.Sync)
	g.POST("/users/resolve", middleware.RequireInternal(), h.Resolve)

	g.GET("/users", middleware.RequirePermission("adaudit.read"), h.List)
	g.GET("/stats", middleware.RequirePermission("adaudit.read"), h.Stats)

	g.POST("/identities", middleware.RequirePermission("adaudit.write"), m.Identity().Create)
	g.GET("/identities", middleware.RequirePermission("adaudit.read"), m.Identity().List)
	g.GET("/identities/:id", middleware.RequirePermission("adaudit.read"), m.Identity().Detail)
	g.DELETE("/identities/:id", middleware.RequirePermission("adaudit.write"), m.Identity().Delete)
	g.PUT("/users/:id/identity", middleware.RequirePermission("adaudit.write"), m.Identity().Assign)
	g.PATCH("/users/:id/identity", middleware.RequirePermission("adaudit.write"), m.Identity().Unassign)
}
