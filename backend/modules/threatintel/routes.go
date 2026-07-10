package threatintel

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/threatintel/handler"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	g := api.Group("/threat-intel", userAuth)

	// POST /api/v1/threat-intel/search → /proxy/api/search/v1/entities/simple
	searchHandler := handler.NewReverseProxyHandler("/proxy/api/search/v1/entities/simple")
	g.POST("/search", searchHandler.Handle)

	// GET /api/v1/threat-intel/entity/:id → /proxy/api/analytics/v1/entity/{id}/details
	g.GET("/entity/:id", func(c *gin.Context) {
		id := c.Param("id")
		targetPath := "/proxy/api/analytics/v1/entity/" + sanitizeId(id) + "/details"
		h := handler.NewReverseProxyHandler(targetPath)
		h.Handle(c)
	})

	// GET /api/v1/threat-intel/entity/:id/relations → /proxy/api/analytics/v1/entity/{id}/relations
	g.GET("/entity/:id/relations", func(c *gin.Context) {
		id := c.Param("id")
		targetPath := "/proxy/api/analytics/v1/entity/" + sanitizeId(id) + "/relations"
		h := handler.NewReverseProxyHandler(targetPath)
		h.Handle(c)
	})

	// GET /api/v1/threat-intel/feeds → /proxy/api/feeds/v1/list
	feedsHandler := handler.NewReverseProxyHandler("/proxy/api/feeds/v1/list")
	g.GET("/feeds", feedsHandler.Handle)

	// POST /api/v1/threat-intel/ai/chat → /proxy/api/ai/v1/chat/completions
	chatHandler := handler.NewReverseProxyHandler("/proxy/api/ai/v1/chat/completions")
	g.POST("/ai/chat", chatHandler.Handle)

	// GET /api/v1/threat-intel/usage → /proxy/usage (special endpoint, no /api prefix)
	usageHandler := handler.NewReverseProxyHandler("/proxy/usage")
	g.GET("/usage", usageHandler.HandleUsageEndpoint)
}

func sanitizeId(id string) string{
	replacer := strings.NewReplacer(".", "", "\\", "","/","")
	return replacer.Replace(id)
}
