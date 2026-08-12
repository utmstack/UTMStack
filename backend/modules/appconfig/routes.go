package appconfig

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc, enterprise gin.HandlerFunc, platform gin.HandlerFunc) {
	// Public, read-only display preference (timezone + date format). Non-sensitive,
	// consumed app-wide (incl. pre-login) to render timestamps consistently.
	api.GET("/date-format", m.Handler().DateFormat)

	g := api.Group("/config", userAuth)
	g.GET("", middleware.RequirePermission("config.read"), m.Handler().List)
	g.GET("/:key", middleware.RequirePermission("config.read"), m.Handler().Get)
	g.PUT("/:key", middleware.RequirePermission("config.write"), m.Handler().Update)
	g.POST("checkEmailConfiguration", middleware.RequirePermission("config.write"), m.Handler().CheckMail)

	b := api.Group("/branding")
	b.GET("", userAuth, middleware.RequirePermission("config.read"), m.BrandingHandler().Get)
	b.PUT("", userAuth, enterprise, middleware.RequirePermission("config.write"), m.BrandingHandler().Update)
	b.POST("/assets/:slot", userAuth, enterprise, middleware.RequirePermission("config.write"), m.BrandingHandler().UploadAsset)
	b.GET("/public", m.BrandingHandler().Public)
	b.POST("/seed", userAuth, middleware.RequireInternal(), m.BrandingHandler().Seed)

	pb := api.Group("/platform/branding", userAuth, platform, middleware.RequirePermission("config.write"), enterprise)
	pb.POST("/bulk/update", m.BulkBrandingHandler().Update)
	pb.POST("/bulk/upload-asset/:slot", m.BulkBrandingHandler().UploadAsset)

	ps := api.Group("/platform/config/smtp", userAuth, platform, middleware.RequirePermission("config.write"))
	ps.POST("/bulk/update", m.BulkSMTPHandler().Update)
	ps.POST("/bulk/test", m.BulkSMTPHandler().Test)
}
