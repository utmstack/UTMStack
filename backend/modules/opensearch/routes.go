package opensearch

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/opensearch/handler"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	propertyH := handler.NewPropertyHandler(m.gateway)
	indexH := handler.NewIndexHandler(m.gateway)
	searchH := handler.NewSearchHandler(m.gateway)
	sqlH := handler.NewSQLHandler(m.gateway)
	clusterH := handler.NewClusterHandler(m.gateway)

	read := middleware.RequirePermission("opensearch.read")
	write := middleware.RequirePermission("opensearch.write")

	g := api.Group("/opensearch", userAuth)

	g.GET("/property/values", read, propertyH.Values)
	g.POST("/property/values-with-count", read, propertyH.ValuesWithCount)
	g.GET("/index/properties", read, indexH.Properties)

	g.POST("/index/delete-index", write, indexH.DeleteIndex)
	g.GET("/index/all", read, indexH.IndexAll)

	g.POST("/search/sql", read, sqlH.SearchSQL)
	g.POST("/search/csv", read, sqlH.SearchCSV)

	g.POST("/search", read, searchH.Search)
	g.POST("/generic-search", read, searchH.GenericSearch)
	g.POST("/count", read, searchH.Count)

	g.GET("/cluster/status", read, clusterH.Status)

	// --- Index pattern catalog (CRUD over utm_index_pattern) ---
	ph := m.GetPatternHandler()
	pg := api.Group("/utm-index-patterns", userAuth)
	pg.POST("", write, ph.Create)
	pg.PUT("", write, ph.Update)
	pg.GET("", read, ph.List)
	pg.GET("/fields", read, ph.Fields) // must be before /:id
	pg.GET("/:id", read, ph.GetByID)
	pg.DELETE("/:id", write, ph.Delete)

	// --- Index lifecycle (ISM) policy — only when OpenSearch is configured ---
	if polH := m.GetPolicyHandler(); polH != nil {
		ipg := api.Group("/index-policy", userAuth)
		ipg.GET("/policy", read, polH.GetPolicy)
		ipg.PUT("/policy", write, polH.UpdatePolicy)
	}
}
