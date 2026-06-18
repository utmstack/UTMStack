package socai

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/socai/handler"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	analyze := handler.NewSocAIHandler(m.client)
	api.POST("/soc-ai/analyze", userAuth, analyze.Analyze)

	g := api.Group("/soc-ai", userAuth)

	// Configuration (admin-only): writes the encrypted plugin YAML.
	cfg := handler.NewConfigHandler(m.config)
	g.GET("/config", middleware.RequireAdmin(), cfg.Get)
	g.PUT("/config", middleware.RequireAdmin(), cfg.Update)

	// Live agent operations (chat-style SSE stream).
	chat := handler.NewChatHandler(m.client)
	g.POST("/chat", chat.Chat)
}
