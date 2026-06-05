package correlation

import (
	"github.com/gin-gonic/gin"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	auditmw "github.com/utmstack/utmstack/backend/modules/audit/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc, auditLogger audit_connectors.Logger) {
	rph := m.GetRegexPatternHandler()
	tch := m.GetTenantConfigHandler()
	dth := m.GetDataTypeHandler()
	crh := m.GetCorrelationRuleHandler()

	// Regex patterns
	rg := api.Group("/regex-pattern", userAuth)
	rg.POST("", rph.Create)
	rg.PUT("", rph.Update)
	rg.GET("", rph.List)
	rg.GET("/:id", rph.GetByID)
	rg.DELETE("/:id", rph.Delete)

	// Tenant config
	tc := api.Group("/tenant-config", userAuth)
	tc.POST("", tch.Create)
	tc.PUT("", tch.Update)
	tc.GET("", tch.List)
	tc.GET("/:id", tch.GetByID)
	tc.DELETE("/:id", tch.Delete)

	// Data types
	dt := api.Group("/data-types", userAuth)
	dt.POST("", dth.Create)
	dt.PUT("", dth.Update)
	dt.PUT("/include-exclude-list", dth.UpdateIncludeExcludeList)
	dt.GET("", dth.List)
	dt.GET("/:id", dth.GetByID)
	dt.DELETE("/:id", dth.Delete)

	// Correlation rules
	cr := api.Group("/correlation-rule", userAuth)

	cr.POST("",
		auditmw.AuditEvent(
			auditLogger,
			audit_domain.CORRELATION_RULE_CREATE_ATTEMPT,
			"Attempting to create correlation rule",
			audit_domain.CORRELATION_RULE_CREATE_SUCCESS,
			"Correlation rule created successfully",
		),
		crh.Create,
	)

	cr.PUT("/activate-deactivate",
		auditmw.AuditEvent(
			auditLogger,
			audit_domain.CORRELATION_RULE_UPDATE_ATTEMPT,
			"Attempting to change activation status for correlation rule",
			audit_domain.CORRELATION_RULE_UPDATE_SUCCESS,
			"Correlation rule activation status changed successfully",
		),
		crh.ActivateDeactivate,
	)

	cr.PUT("",
		auditmw.AuditEvent(
			auditLogger,
			audit_domain.CORRELATION_RULE_UPDATE_ATTEMPT,
			"Attempting to update correlation rule",
			audit_domain.CORRELATION_RULE_UPDATE_SUCCESS,
			"Correlation rule updated successfully",
		),
		crh.Update,
	)

	cr.GET("/search-by-filters", crh.List)

	cr.GET("/search-property-values", crh.SearchPropertyValues)

	cr.GET("/:id", crh.GetByID)

	cr.DELETE("/:id",
		auditmw.AuditEvent(
			auditLogger,
			audit_domain.CORRELATION_RULE_DELETE_ATTEMPT,
			"Attempting to delete correlation rule",
			audit_domain.CORRELATION_RULE_DELETE_SUCCESS,
			"Correlation rule deleted successfully",
		),
		crh.Delete,
	)
}
