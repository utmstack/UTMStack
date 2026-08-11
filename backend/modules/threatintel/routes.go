package threatintel

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/threatintel/handler"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
	"github.com/utmstack/utmstack/backend/pkg/instanceconfig"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	g := api.Group("/threat-intel", userAuth, middleware.RequirePermission("threatintel.read"))

	// POST /api/v1/threat-intel/search → /proxy/api/search/v1/entities/simple
	searchHandler := handler.NewReverseProxyHandler("/proxy/api/search/v1/entities/simple")
	g.POST("/search", searchHandler.Handle)

	// POST /api/v1/threat-intel/search/advanced → /proxy/api/search/v1/entities/advanced
	advancedSearchHandler := handler.NewReverseProxyHandler("/proxy/api/search/v1/entities/advanced")
	g.POST("/search/advanced", advancedSearchHandler.Handle)

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

	// GET /api/v1/threat-intel/status — can this instance reach ThreatWinds at
	// all. Answers a boolean and nothing else, so the page can say "not set up"
	// without the caller being told what the instance is spending. It replaces
	// probing /usage for the same answer, which also made a proxy round trip to
	// learn something already known locally.
	g.GET("/status", func(c *gin.Context) {
		cfg := instanceconfig.Get()
		configured := cfg != nil && cfg.Server != "" && cfg.InstanceID != "" && cfg.InstanceKey != ""
		c.JSON(http.StatusOK, gin.H{"configured": configured})
	})

	// GET /api/v1/threat-intel/usage → /proxy/usage (special endpoint, no /api prefix)
	//
	// Platform-only. Everything else here proxies public threat intelligence,
	// which is the same for everyone, but this counter is the instance's own
	// consumption against its ThreatWinds plan — summed over every tenant on it.
	// Handing it to a tenant tells them how much the others are using. Their own
	// allowance is a different number, and they read it from /soc-ai/usage.
	usageHandler := handler.NewReverseProxyHandler("/proxy/usage")
	g.GET("/usage", middleware.RequirePlatform(), usageHandler.HandleUsageEndpoint)

	// Contributing this instance's incidents back to ThreatWinds. The switch is
	// the platform's — it decides what leaves the install — and the credentials
	// are written by the plugin that registered them, never read back out.
	if fh := m.FeedsHandler(); fh != nil {
		g.GET("/feeds/contribution", middleware.RequirePlatform(), fh.Status)
		g.PUT("/feeds/contribution", middleware.RequirePlatform(), middleware.RequireAdmin(), fh.SetEnabled)
		g.PUT("/feeds/credentials", middleware.RequireInternal(), fh.SaveCredentials)
	}
}

func sanitizeId(id string) string {
	replacer := strings.NewReplacer(".", "", "\\", "", "/", "")
	return replacer.Replace(id)
}
