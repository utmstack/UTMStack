package datasources

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	ds := m.GetDatasourceHandler()
	ck := m.GetConnectionKeyHandler()

	d := api.Group("/datasources", userAuth)
	d.GET("/count", middleware.RequirePermission("datasources.read"), ds.Count)
	d.GET("/connection-key", middleware.RequireAdmin(), ck.Get)
	d.POST("/connection-key/rotate", middleware.RequireAdmin(), ck.Rotate)
	d.GET("", middleware.RequirePermission("datasources.read"), ds.List)
	d.GET("/:id", middleware.RequirePermission("datasources.read"), ds.Get)
	d.PUT("/labels", middleware.RequirePermission("datasources.write"), ds.UpdateLabels)
	d.PUT("/sensitivity", middleware.RequirePermission("datasources.write"), ds.UpdateSensitivity)
	d.DELETE("/:id", middleware.RequirePermission("datasources.write"), ds.Delete)

	d.POST("/ping", middleware.RequireInternal(), ds.Ping)
}
