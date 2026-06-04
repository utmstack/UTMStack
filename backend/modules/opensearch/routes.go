package opensearch

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/opensearch/handler"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	es := api.Group("/elasticsearch", userAuth)

	propertyH := handler.NewPropertyHandler(m.gateway)
	indexH := handler.NewIndexHandler(m.gateway)
	searchH := handler.NewSearchHandler(m.gateway)
	sqlH := handler.NewSQLHandler(m.gateway)
	clusterH := handler.NewClusterHandler(m.gateway)

	es.GET("/property/values", propertyH.Values)
	es.POST("/property/values-with-count", propertyH.ValuesWithCount)
	es.GET("/index/properties", indexH.Properties)

	es.POST("/index/delete-index", indexH.DeleteIndex)
	es.GET("/index/all", indexH.IndexAll)

	es.POST("/search/sql", sqlH.SearchSQL)
	es.POST("/search/csv", sqlH.SearchCSV)

	es.POST("/search", searchH.Search)
	es.POST("/generic-search", searchH.GenericSearch)
	es.POST("/count", searchH.Count)

	es.GET("/cluster/status", clusterH.Status)
}
