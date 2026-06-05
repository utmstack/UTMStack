package integrations

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/integrations/handler"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	mh := handler.NewModuleHandler(m.Modules())
	th := handler.NewTenantHandler(m.Tenants())

	read := middleware.RequirePermission("integrations.read")
	write := middleware.RequirePermission("integrations.write")

	g := api.Group("/integrations", userAuth)

	g.GET("", read, mh.List)
	g.GET("/categories", read, mh.Categories)
	g.GET("/is-active", read, mh.IsActive)
	g.PUT("/activate", write, mh.ActivateDeactivate)
	g.GET("/:id", read, mh.Get)

	g.POST("", write, mh.Create)
	g.PUT("/:id", write, mh.Update)
	g.DELETE("/:id", write, mh.Delete)

	g.GET("/tenants/:module", read, th.List)
	g.PUT("/tenants/:module", write, th.Save)
	g.DELETE("/tenants/:module/:name", write, th.Delete)

	g.GET("/tenants/:module/config", th.PluginConfig)
}
