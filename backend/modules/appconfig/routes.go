package appconfig

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc, mssp gin.HandlerFunc) {
	g := api.Group("/config", userAuth)
	g.GET("", middleware.RequirePermission("config.read"), m.Handler().List)
	g.GET("/:key", middleware.RequirePermission("config.read"), m.Handler().Get)
	g.PUT("/:key", middleware.RequirePermission("config.write"), m.Handler().Update)
	g.POST("checkEmailConfiguration", middleware.RequirePermission("config.write"), m.Handler().CheckMail)

	b := api.Group("/branding")
	b.GET("", userAuth, middleware.RequirePermission("config.read"), m.BrandingHandler().Get)
	b.PUT("", userAuth, mssp, middleware.RequirePermission("config.write"), m.BrandingHandler().Update)
	b.GET("/public", m.BrandingHandler().Public)
}
