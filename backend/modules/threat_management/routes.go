package threat_management

import "github.com/gin-gonic/gin"

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	adversary := api.Group("/adversary")
	adversary.Use(userAuth)
	{
		adversary.POST("/alerts", m.adversaryHandler.SearchAdversaryAlerts)
	}
}
