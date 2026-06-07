package dashboards

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	dh := m.GetDashboardHandler()
	vh := m.GetVisualizationHandler()
	lh := m.GetLayoutHandler()

	read := middleware.RequirePermission("dashboards.read")
	write := middleware.RequirePermission("dashboards.write")

	d := api.Group("/dashboards", userAuth)
	d.POST("", write, dh.Create)
	d.PUT("", write, dh.Update)
	d.GET("", read, dh.List)
	d.GET("/:id", read, dh.GetByID)
	d.DELETE("/:id", write, dh.Delete)

	v := api.Group("/visualizations", userAuth)
	v.POST("", write, vh.Create)
	v.PUT("", write, vh.Update)
	v.GET("", read, vh.List)
	v.GET("/:id", read, vh.GetByID)
	v.DELETE("/:id", write, vh.Delete)

	l := api.Group("/dashboard-layouts", userAuth)
	l.POST("", write, lh.Create)
	l.PUT("", write, lh.Update)
	l.GET("", read, lh.List)
	l.GET("/:id", read, lh.GetByID)
	l.DELETE("/:id", write, lh.Delete)
}
