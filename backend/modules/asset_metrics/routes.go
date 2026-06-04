package asset_metrics

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module) {
	h := m.handler

	rg := api.Group("/utm-asset-metrics")
	rg.POST("", h.Create)
	rg.PUT("", h.Update)
	rg.GET("", h.ListAll)
	rg.GET("/:id", h.GetByID)
	rg.DELETE("/:id", h.Delete)
}
