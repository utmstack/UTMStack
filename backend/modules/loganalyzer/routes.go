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
	g.POST("/top-x-values/:dataset/:field/:top", read, ah.TopValues)

	// What can be explored, and what can be asked of it. Both used to come from
	// the index-pattern registry an operator maintained; the event store has two
	// datasets and knows its own columns.
	g.GET("/datasets", read, ah.Datasets)
	g.GET("/datasets/:dataset/fields", read, ah.Fields)
	g.GET("/datasets/:dataset/data-types", read, ah.DataTypes)
	g.POST("/chart-view", read, ah.ChartView)
	g.POST("/search", read, ah.Search)

	q := g.Group("/queries")
	q.POST("", write, qh.Create)
	q.PUT("", write, qh.Update)
	q.GET("", read, qh.List)
	q.GET("/:id", read, qh.GetByID)
	q.DELETE("/:id", write, qh.Delete)
}
