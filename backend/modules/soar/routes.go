package soar

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	rh := m.GetRuleHandler()
	th := m.GetTemplateHandler()
	eh := m.GetExecutionHandler()

	g := api.Group("/soar", userAuth)

	// Response rules
	rg := g.Group("/rules")
	rg.POST("", middleware.RequirePermission("soar.write"), rh.Create)
	rg.PUT("", middleware.RequirePermission("soar.write"), rh.Update)
	rg.GET("", middleware.RequirePermission("soar.read"), rh.List)
	rg.GET("/resolve-filter-values", middleware.RequirePermission("soar.read"), rh.ResolveFilterValues) // BEFORE /:id
	rg.GET("/:id", middleware.RequirePermission("soar.read"), rh.GetByID)

	// Reusable action templates
	g.GET("/action-templates", middleware.RequirePermission("soar.read"), th.List)

	// Rule executions
	g.GET("/rule-executions", middleware.RequirePermission("soar.read"), eh.List)
}
