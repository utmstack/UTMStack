package socai

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/socai/handler"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	h := handler.NewSocAIHandler(m.client)
	api.POST("/soc-ai/analyze", userAuth, h.Analyze)
}
