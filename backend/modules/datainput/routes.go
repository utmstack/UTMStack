package datainput

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	h := m.GetHandler()

	rg := api.Group("/utm-data-input-statuses", userAuth)
	rg.POST("", h.Create)
	rg.PUT("", h.Update)
	rg.GET("", h.List)
	rg.GET("/count", h.Count)
	rg.GET("/:id", h.GetByID)
	rg.DELETE("/:id", h.Delete)
}
