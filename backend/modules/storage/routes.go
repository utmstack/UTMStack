package storage

import (
	"github.com/gin-gonic/gin"

	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	h := m.Handler()
	if h == nil {
		return
	}

	read := middleware.RequirePermission("storage.read")
	write := middleware.RequirePermission("storage.write")

	g := api.Group("/storage", userAuth, middleware.RequirePlatform())

	g.GET("/retention", read, h.Retentions)
	g.PUT("/retention", write, h.SetRetention)

	g.GET("/usage", read, h.Usage)
	g.GET("/health", read, h.Health)

	g.GET("/tiering", read, h.Tiering)
	g.PUT("/tiering", write, h.EnableTiering)
}
