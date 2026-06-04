package logstash

import (
	"github.com/gin-gonic/gin"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	auditmw "github.com/utmstack/utmstack/backend/modules/audit/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc, auditLogger audit_connectors.Logger) {
	fgHandler := m.GetFilterGroupHandler()
	fg := api.Group("/utm-logstash-filter-groups", userAuth)
	fg.POST("", fgHandler.Create)
	fg.PUT("", fgHandler.Update)
	fg.GET("", fgHandler.List)
	fg.GET("/count", fgHandler.Count)
	fg.GET("/:id", fgHandler.GetByID)
	fg.DELETE("/:id", fgHandler.Delete)

	fHandler := m.GetFilterHandler()
	f := api.Group("/utm-filters", userAuth)
	f.POST("",
		auditmw.AuditEvent(
			auditLogger,
			audit_domain.LOGSTASH_FILTER_CREATE_ATTEMPT, "logstash filter create attempt",
			audit_domain.LOGSTASH_FILTER_CREATE_SUCCESS, "logstash filter created",
		),
		fHandler.Create,
	)
	f.PUT("",
		auditmw.AuditEvent(
			auditLogger,
			audit_domain.LOGSTASH_FILTER_UPDATE_ATTEMPT, "logstash filter update attempt",
			audit_domain.LOGSTASH_FILTER_UPDATE_SUCCESS, "logstash filter updated",
		),
		fHandler.Update,
	)
	f.GET("", fHandler.List)
	f.GET("/by-pipelineid", fHandler.FiltersByPipelineID) // BEFORE /:id
	f.GET("/:id", fHandler.GetByID)
	f.DELETE("/:id",
		auditmw.AuditEvent(
			auditLogger,
			audit_domain.LOGSTASH_FILTER_DELETE_ATTEMPT, "logstash filter delete attempt",
			audit_domain.LOGSTASH_FILTER_DELETE_SUCCESS, "logstash filter deleted",
		),
		fHandler.Delete,
	)

	pHandler := m.GetPipelineHandler()
	p := api.Group("/logstash-pipelines", userAuth)
	p.GET("", pHandler.List)
	p.GET("/stats", pHandler.GetStats)     // before /:id
	p.POST("/validate", pHandler.Validate) // before /:id
	p.GET("/:id", pHandler.GetByID)
	p.DELETE("/:id", pHandler.Delete)
}
