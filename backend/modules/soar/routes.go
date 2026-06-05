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

	// Incident variables: reusable (optionally secret) values interpolated into
	// commands run on agents. Kept on the legacy /utm-incident-variables path so
	// the existing console keeps working.
	vh := m.GetVariableHandler()
	vg := api.Group("/utm-incident-variables", userAuth)
	vg.POST("", write, vh.Create)
	vg.PUT("", write, vh.Update)
	vg.GET("", read, vh.List)
	vg.GET("/:id", read, vh.GetByID)
	vg.DELETE("/:id", write, vh.Delete)

	// Incident-response automation: predefined actions and their per-OS commands.
	ah := m.GetActionHandler()
	ag := api.Group("/utm-incident-actions", userAuth)
	ag.POST("", write, ah.Create)
	ag.PUT("", write, ah.Update)
	ag.GET("", read, ah.List)
	ag.GET("/count", read, ah.Count) // BEFORE /:id
	ag.GET("/:id", read, ah.GetByID)
	ag.DELETE("/:id", write, ah.Delete)

	ach := m.GetActionCommandHandler()
	acg := api.Group("/incident-action-commands", userAuth)
	acg.POST("", write, ach.Create)
	acg.PUT("", write, ach.Update)
	acg.GET("", read, ach.List)
	acg.GET("/:id", read, ach.GetByID)
	acg.DELETE("/:id", write, ach.Delete)

	// Incident jobs: responses run against agents, surfaced in incident reports.
	jh := m.GetJobHandler()
	jg := api.Group("/utm-incident-jobs", userAuth)
	jg.POST("", write, jh.Create)
	jg.GET("", read, jh.List)
	jg.GET("/count", read, jh.Count) // BEFORE /:id
	jg.GET("/:id", read, jh.GetByID)
	jg.DELETE("/:id", write, jh.Delete)

	// Live command execution (interactive console).
	api.GET("/soar/ws/command/:hostname", m.commandWSHandler.CommandStream)
}
