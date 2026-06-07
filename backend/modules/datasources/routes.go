package datasources

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	ds := m.GetDatasourceHandler()
	gr := m.GetAssetGroupHandler()

	d := api.Group("/datasources", userAuth)
	d.GET("", middleware.RequirePermission("datasources.read"), ds.List)
	d.GET("/:id", middleware.RequirePermission("datasources.read"), ds.Get)
	d.PUT("/group", middleware.RequirePermission("datasources.write"), ds.UpdateGroup)
	d.PUT("/labels", middleware.RequirePermission("datasources.write"), ds.UpdateLabels)
	d.DELETE("/:id", middleware.RequirePermission("datasources.write"), ds.Delete)

	d.POST("/ping", middleware.RequireInternal(), ds.Ping)

	g := api.Group("/datasource-groups", userAuth)
	g.GET("", middleware.RequirePermission("datasources.read"), gr.List)
	g.POST("", middleware.RequirePermission("datasources.write"), gr.Create)
	g.GET("/:id", middleware.RequirePermission("datasources.read"), gr.Get)
	g.PUT("/:id", middleware.RequirePermission("datasources.write"), gr.Update)
	g.DELETE("/:id", middleware.RequirePermission("datasources.write"), gr.Delete)
}
