package billing

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/billing/handler"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	vh := handler.NewVersionHandler(m.Version(), m.License())
	lh := handler.NewLicenseHandler(m.License())

	g := api.Group("/billing", userAuth)
	g.GET("/version", vh.Get)
	g.GET("/license", lh.Get)
	g.POST("/license", middleware.RequireAdmin(), lh.Upload) // upload/replace the LICENSE
}
