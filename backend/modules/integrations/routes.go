package integrations

import (
	"github.com/gin-gonic/gin"

	"github.com/utmstack/utmstack/backend/modules/integrations/handler"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	ih := handler.NewIntegrationHandler(m.Integrations())
	gh := handler.NewConfigGroupHandler(m.Groups())
	ch := handler.NewCollectorHandler(m.Collectors())

	read := middleware.RequirePermission("integrations.read")
	write := middleware.RequirePermission("integrations.write")

	g := api.Group("/integrations", userAuth)

	g.GET("", read, ih.List)
	g.GET("/data-types", read, ih.DataTypes)
	g.GET("/:id", read, ih.Get)

	g.POST("", write, ih.Create)
	g.PUT("/:id", write, ih.Update)
	g.DELETE("/:id", write, ih.Delete)

	g.GET("/config/:integration", read, gh.List)
	g.PUT("/config/:integration", write, gh.Save)
	g.DELETE("/config/:integration/:name", write, gh.Delete)

	g.GET("/collectors", write, ch.ListForwarders)
	g.GET("/collectors/:id/data-types/:dataType", write, ch.GetDataType)
	g.PUT("/collectors/:id/data-types/:dataType", write, ch.SetDataType)
	g.PUT("/collectors/:id/certificates", write, ch.SetCertificates)
	g.GET("/collectors/:id/tls-status", write, ch.GetTLSStatus)
}
