package loganalyzer

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	qh := m.GetQueryHandler()
	ah := m.GetAnalyzerHandler()

	read := middleware.RequirePermission("loganalyzer.read")
	write := middleware.RequirePermission("loganalyzer.write")

	g := api.Group("/log-analyzer", userAuth)

	// Ad-hoc aggregations (read).
	g.POST("/top-x-values/:indexPattern/:field/:top", read, ah.TopValues)
	g.POST("/chart-view", read, ah.ChartView)

	q := g.Group("/queries")
	q.POST("", write, qh.Create)
	q.PUT("", write, qh.Update)
	q.GET("", read, qh.List)
	q.GET("/:id", read, qh.GetByID)
	q.DELETE("/:id", write, qh.Delete)
}
