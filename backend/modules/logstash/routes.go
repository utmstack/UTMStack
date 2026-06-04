package logstash

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	fgHandler := m.GetFilterGroupHandler()
	fg := api.Group("/utm-logstash-filter-groups", userAuth)
	fg.POST("", fgHandler.Create)
	fg.PUT("", fgHandler.Update)
	fg.GET("", fgHandler.List)
	fg.GET("/count", fgHandler.Count)
	fg.GET("/:id", fgHandler.GetByID)
	fg.DELETE("/:id", fgHandler.Delete)

	// Filters — audit is recorded handler-side via audit.Record.
	fHandler := m.GetFilterHandler()
	f := api.Group("/utm-filters", userAuth)
	f.POST("", fHandler.Create)
	f.PUT("", fHandler.Update)
	f.GET("", fHandler.List)
	f.GET("/by-pipelineid", fHandler.FiltersByPipelineID) // BEFORE /:id
	f.GET("/:id", fHandler.GetByID)
	f.DELETE("/:id", fHandler.Delete)

	pHandler := m.GetPipelineHandler()
	p := api.Group("/logstash-pipelines", userAuth)
	p.GET("", pHandler.List)
	p.GET("/stats", pHandler.GetStats)     // before /:id
	p.POST("/validate", pHandler.Validate) // before /:id
	p.GET("/:id", pHandler.GetByID)
	p.DELETE("/:id", pHandler.Delete)
}
