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
}
