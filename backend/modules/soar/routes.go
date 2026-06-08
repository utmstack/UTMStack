package soar

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	rh := m.GetRuleHandler()
	th := m.GetTemplateHandler()
	eh := m.GetExecutionHandler()
	agh := m.GetAgentHandler()
	vh := m.GetVariableHandler()
	ah := m.GetActionHandler()
	ach := m.GetActionCommandHandler()
	jh := m.GetJobHandler()

	read := middleware.RequirePermission("soar.read")
	write := middleware.RequirePermission("soar.write")

	g := api.Group("/soar", userAuth)

	rg := g.Group("/rules")
	rg.POST("", write, rh.Create)
	rg.GET("", read, rh.List)
	rg.GET("/resolve-filter-values", read, rh.ResolveFilterValues)
	rg.GET("/:relPath", read, rh.Get)
	rg.PUT("/:relPath", write, rh.Update)
	rg.DELETE("/:relPath", write, rh.Delete)
	rg.PUT("/:relPath/enabled", write, rh.SetEnabled)

	g.GET("/action-templates", read, th.List)

	g.GET("/rule-executions", read, eh.List)
	// Internal-only writes for the SOAR plugin runtime.
	g.POST("/rule-executions", middleware.RequireInternal(), eh.Create)
	g.PATCH("/rule-executions/:id", middleware.RequireInternal(), eh.UpdateStatus)
	g.GET("/agents", middleware.RequireInternal(), agh.List)

	vg := g.Group("/incident-variables")
	vg.POST("", write, vh.Create)
	vg.PUT("", write, vh.Update)
	vg.GET("", read, vh.List)
	vg.GET("/:id", read, vh.GetByID)
	vg.DELETE("/:id", write, vh.Delete)

	ag := g.Group("/incident-actions")
	ag.POST("", write, ah.Create)
	ag.PUT("", write, ah.Update)
	ag.GET("", read, ah.List)
	ag.GET("/count", read, ah.Count)
	ag.GET("/:id", read, ah.GetByID)
	ag.DELETE("/:id", write, ah.Delete)

	acg := g.Group("/incident-action-commands")
	acg.POST("", write, ach.Create)
	acg.PUT("", write, ach.Update)
	acg.GET("", read, ach.List)
	acg.GET("/:id", read, ach.GetByID)
	acg.DELETE("/:id", write, ach.Delete)

	jg := g.Group("/incident-jobs")
	jg.POST("", write, jh.Create)
	jg.GET("", read, jh.List)
	jg.GET("/count", read, jh.Count)
	jg.GET("/:id", read, jh.GetByID)
	jg.DELETE("/:id", write, jh.Delete)

	api.GET("/soar/ws/command/:agentId", m.commandWSHandler.CommandStream)
}
