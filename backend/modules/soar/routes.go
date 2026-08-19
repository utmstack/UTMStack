package soar

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc, apiKeyAuth middleware.APIKeyAuthFunc, platform gin.HandlerFunc) {
	rh := m.GetRuleHandler()
	eh := m.GetExecutionHandler()
	vh := m.GetVariableHandler()
	bh := m.GetBulkHandler()

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

	g.GET("/rule-executions", read, eh.List)
	g.POST("/rule-executions", middleware.RequireInternal(), eh.Match)

	vg := g.Group("/variables")
	vg.POST("", write, vh.Create)
	vg.PUT("", write, vh.Update)
	vg.GET("", read, vh.List)
	vg.GET("/:id", read, vh.GetByID)
	vg.DELETE("/:id", write, vh.Delete)

	m.commandWSHandler.SetAPIKeyAuth(apiKeyAuth)
	api.GET("/soar/ws/command/:agentId", m.commandWSHandler.CommandStream)

	// Platform-admin bulk endpoints (rules only).
	prg := api.Group("/platform/soar/rules/bulk", userAuth, platform, write)
	prg.POST("/create", bh.BulkCreateRule)
	prg.POST("/update", bh.BulkUpdateRule)
	prg.POST("/delete", bh.BulkDeleteRule)
	prg.POST("/enable", bh.BulkEnableRule)
}
