package audit

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	g := api.Group("/audit", userAuth)
	g.GET("", middleware.RequirePermission("audit.read"), m.Handler().List)
	g.GET("/:id", middleware.RequirePermission("audit.read"), m.Handler().Get)
}
