package soar

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	rh := m.GetRuleHandler()
	th := m.GetTemplateHandler()
	eh := m.GetExecutionHandler()

	read := middleware.RequirePermission("soar.read")
	write := middleware.RequirePermission("soar.write")

	g := api.Group("/soar", userAuth)

	// Response rules
	rg := g.Group("/rules")
	rg.POST("", write, rh.Create)
	rg.PUT("", write, rh.Update)
	rg.GET("", read, rh.List)
	rg.GET("/resolve-filter-values", read, rh.ResolveFilterValues) // BEFORE /:id
	rg.GET("/:id", read, rh.GetByID)

	// Reusable action templates
	g.GET("/action-templates", read, th.List)

	// Rule executions
	g.GET("/rule-executions", read, eh.List)

	// Live command execution (interactive console).
	api.GET("/soar/ws/command/:hostname", m.commandWSHandler.CommandStream)
}
