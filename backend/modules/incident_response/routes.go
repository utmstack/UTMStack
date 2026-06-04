package incident_response

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	// Incident actions
	ag := api.Group("/utm-incident-actions", userAuth)
	ag.POST("", m.actionHandler.Create)
	ag.PUT("", m.actionHandler.Update)
	ag.GET("", m.actionHandler.List)
	ag.GET("/count", m.actionHandler.Count)
	ag.GET("/:id", m.actionHandler.GetByID)
	ag.DELETE("/:id", m.actionHandler.Delete)

	// Incident action commands
	cg := api.Group("/incident-action-commands", userAuth)
	cg.POST("", m.actionCommandHandler.Create)
	cg.PUT("", m.actionCommandHandler.Update)
	cg.GET("", m.actionCommandHandler.List)
	cg.GET("/:id", m.actionCommandHandler.GetByID)
	cg.DELETE("/:id", m.actionCommandHandler.Delete)

	// Incident jobs
	jg := api.Group("/utm-incident-jobs", userAuth)
	jg.POST("", m.jobHandler.Create)
	jg.GET("", m.jobHandler.List)
	jg.GET("/count", m.jobHandler.Count)
	jg.GET("/:id", m.jobHandler.GetByID)
	jg.DELETE("/:id", m.jobHandler.Delete)

	// Incident variables
	vg := api.Group("/utm-incident-variables", userAuth)
	vg.POST("", m.variableHandler.Create)
	vg.PUT("", m.variableHandler.Update)
	vg.GET("", m.variableHandler.List)
	vg.GET("/:id", m.variableHandler.GetByID)
	vg.DELETE("/:id", m.variableHandler.Delete)

	// WebSocket: no userAuth at router level — JWT validated inside handler.
	api.GET("/ws/incident-command/:hostname", m.commandWSHandler.CommandStream)
}
