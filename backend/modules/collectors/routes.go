package collectors

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	collectors := api.Group("/collectors", userAuth)
	collectors.GET("", m.collectorHandler.ListCollectors)
	collectors.DELETE("/:id", m.collectorHandler.DeleteCollector)

	agents := api.Group("/agent-manager", userAuth)
	agents.GET("/agents", m.agentHandler.ListAgents)
	agents.GET("/agents-with-commands", m.agentHandler.ListAgentsWithCommands)
	agents.GET("/agent-commands", m.agentHandler.ListAgentCommands)
	agents.GET("/agent-by-hostname", m.agentHandler.GetAgentByHostname)
	agents.GET("/can-run-command", m.agentHandler.CanRunCommand)
	agents.POST("/update-agent-attrs", m.agentHandler.UpdateAgentAttrs)
}
