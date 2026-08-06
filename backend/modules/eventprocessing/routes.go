package eventprocessing

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	rph := m.GetRegexPatternHandler()
	crh := m.GetCorrelationRuleHandler()
	fh := m.GetFilterHandler()
	sh := m.GetIngestionStatsHandler()

	read := middleware.RequirePermission("eventprocessing.read")
	write := middleware.RequirePermission("eventprocessing.write")

	g := api.Group("/eventprocessing", userAuth)

	// Regex patterns are read-only. They are a shared vocabulary referenced from
	// filter YAMLs as {{.name}}, seeded by the pipeline bootstrap and identical
	// for every tenant — consumed here, never authored.
	rg := g.Group("/regex-pattern")
	rg.GET("", read, rph.List)
	rg.GET("/:id", read, rph.GetByID)

	// Tenant config (assets) is retired: asset CIA now lives on datasources, which
	// project it into tenants.yaml via the eventprocessing TenantConfig usecase.

	// Correlation rules — audit is recorded handler-side via audit.Record.
	cr := g.Group("/correlation-rule")
	cr.POST("", write, crh.Create)
	cr.POST("/import", write, crh.Import)
	cr.POST("/export", read, crh.Export)
	cr.PUT("/activate-deactivate", write, crh.ActivateDeactivate)
	cr.PUT("", write, crh.Update)
	cr.GET("/search-by-filters", read, crh.List)
	cr.GET("/search-property-values", read, crh.SearchPropertyValues)
	cr.GET("/find", read, crh.GetByID)
	cr.DELETE("", write, crh.Delete)

	// Filters (file-backed, pipeline: YAML). Identity = relPath query param.
	f := g.Group("/filters")
	f.POST("", write, fh.Create)
	f.PUT("", write, fh.Update)
	f.PUT("/activate", write, fh.ActivateDeactivate) // BEFORE /find
	f.PUT("/order", write, fh.SetOrder)              // reorder a custom filter in the global pipeline order
	f.GET("", read, fh.List)
	f.GET("/data-types", read, fh.DataTypes) // distinct dataTypes for the UI filter
	f.GET("/find", read, fh.GetByRelPath)    // BEFORE potential /:id
	f.DELETE("", write, fh.Delete)

	// Live ingestion stats (read from v11-statistics-*; no DB/sync).
	st := g.Group("/ingestion-stats")
	st.GET("", read, sh.Totals)
	st.GET("/timeline", read, sh.Timeline)

	ph := m.GetPlaygroundHandler()
	pg := g.Group("/playground")
	pg.POST("/test-filter", write, ph.TestFilter)
	pg.POST("/test-rule", write, ph.TestRule)
}
