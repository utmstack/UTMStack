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

	cfg := handler.NewConfigHandler(m.config)
	g.GET("/config", middleware.RequireAdmin(), cfg.Get)
	g.PUT("/config", middleware.RequireAdmin(), cfg.Update)
	g.DELETE("/config", middleware.RequireAdmin(), cfg.ResetToDefault)

	// Live agent operations (chat-style SSE stream).
	chat := handler.NewChatHandler(m.client)
	g.POST("/chat", chat.Chat)
}
