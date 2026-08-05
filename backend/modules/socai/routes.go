package socai

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/socai/handler"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	use := middleware.RequirePermission("socai.read")
	configure := middleware.RequirePermission("socai.write")

	quota := m.quota.Gate()

	analyze := handler.NewSocAIHandler(m.client)
	api.POST("/soc-ai/analyze", userAuth, use, quota, analyze.Analyze)

	g := api.Group("/soc-ai", userAuth)

	cfg := handler.NewConfigHandler(m.config)
	g.GET("/config", configure, cfg.Get)
	g.PUT("/config", configure, cfg.Update)
	g.DELETE("/config", configure, cfg.ResetToDefault)

	if m.quota != nil && m.quota.LimitOf != nil && m.quota.Used != nil {
		usage := handler.NewUsageHandler(m.quota.LimitOf, m.quota.Used)
		g.GET("/usage", use, usage.Usage)
	}

	// Live agent operations (chat-style SSE stream).
	chat := handler.NewChatHandler(m.client)
	g.POST("/chat", use, quota, chat.Chat)

	// Alerts reach the plugin straight from the pipeline, never through here, so
	// automatic analysis would spend outside every limit. The plugin asks first,
	// and asking is the same gate the two routes above use — one counter, one
	// set of rules, whichever way the request arrived.
	ig := api.Group("/soc-ai", userAuth, middleware.RequireInternal())
	ig.POST("/quota/consume", quota, func(c *gin.Context) { c.Status(http.StatusNoContent) })
}
