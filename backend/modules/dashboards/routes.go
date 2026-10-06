package dashboards

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	dh := m.GetDashboardHandler()
	vh := m.GetVisualizationHandler()
	fh := m.GetDashboardFilterHandler()

	read := middleware.RequirePermission("dashboards.read")
	write := middleware.RequirePermission("dashboards.write")

	d := api.Group("/dashboards", userAuth)
	d.POST("", write, dh.Create)
	d.PUT("", write, dh.Update)
	d.GET("", read, dh.List)
	d.GET("/:id", read, dh.GetByID)
	d.DELETE("/:id", write, dh.Delete)

	f := d.Group("/filters", userAuth)
	f.POST("", write, fh.Create)
	f.PUT("", write, fh.Update)
	f.GET("", read, fh.List)
	f.DELETE("/:id", write, fh.Delete)

	v := api.Group("/visualizations", userAuth)
	v.POST("", write, vh.Create)
	v.PUT("", write, vh.Update)
	v.GET("", read, vh.List)
	v.GET("/:id", read, vh.GetByID)
	v.DELETE("/:id", write, vh.Delete)

	// Answering a widget is a read of the event store, and the spec cannot name
	// a tenant: every call reaches it through a scope carrying the caller's own.
	if qh := m.QueryHandler(); qh != nil {
		v.POST("/query", read, qh.Run)
	}
}
