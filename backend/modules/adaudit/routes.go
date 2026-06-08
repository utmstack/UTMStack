package adaudit

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	h := m.Handler()
	g := api.Group("/ad-audit", userAuth)

	// Internal: the ad-audit plugin pushes changed users and seeds its cache.
	g.POST("/users", middleware.RequireInternal(), h.Ingest)
	g.GET("/users/sync", middleware.RequireInternal(), h.Sync)

	// UI: list the AD user inventory.
	g.GET("/users", middleware.RequirePermission("adaudit.read"), h.List)
}
