package indexpattern

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	// --- Pattern CRUD ---
	h := m.GetHandler()
	rg := api.Group("/utm-index-patterns", userAuth)
	rg.POST("", h.Create)
	rg.PUT("", h.Update)
	rg.GET("", h.List)
	rg.GET("/fields", h.Fields) // must be before /:id
	rg.GET("/:id", h.GetByID)
	rg.DELETE("/:id", h.Delete)

	ph := m.GetPolicyHandler()
	if ph != nil {
		pg := api.Group("/index-policy", userAuth)
		pg.GET("/policy", ph.GetPolicy)
		pg.PUT("/policy", ph.UpdatePolicy)
	}
}
