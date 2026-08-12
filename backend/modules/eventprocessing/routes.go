package eventprocessing

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/handler"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc, platform gin.HandlerFunc, tenantLister func(context.Context) ([]string, error)) {
	rph := m.GetRegexPatternHandler()
	crh := m.GetCorrelationRuleHandler()
	fh := m.GetPipelineHandler()
	sh := m.GetIngestionStatsHandler()

	read := middleware.RequirePermission("eventprocessing.read")
	write := middleware.RequirePermission("eventprocessing.write")

	g := api.Group("/eventprocessing", userAuth)

	rg := g.Group("/regex-pattern")
	rg.GET("", read, rph.List)
	rg.GET("/:id", read, rph.GetByID)

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

	// Pipelines (file-backed YAML). Identity = relPath query param.
	f := g.Group("/pipelines")
	f.POST("", write, fh.Create)
	f.PUT("", write, fh.Update)
	f.PUT("/activate", write, fh.ActivateDeactivate) // BEFORE /find
	f.PUT("/order", write, fh.SetOrder)              // reorder a pipeline against the others matching its data type
	f.GET("", read, fh.List)
	f.GET("/data-types", read, fh.DataTypes) // distinct dataTypes for the UI selector
	f.GET("/find", read, fh.GetByRelPath)    // BEFORE potential /:id
	f.DELETE("", write, fh.Delete)

	// Live ingestion stats, read from the event store.
	st := g.Group("/ingestion-stats")
	st.GET("", read, sh.Totals)
	st.GET("/timeline", read, sh.Timeline)

	ph := m.GetPlaygroundHandler()
	pg := g.Group("/playground")
	pg.POST("/test-pipeline", write, ph.TestPipeline)
	pg.POST("/test-rule", write, ph.TestRule)

	// Platform-admin bulk endpoints (default-tenant admins only).
	bh := handler.NewBulkPipelineHandler(m.GetPipelineUsecase(), tenantLister)
	bp := api.Group("/platform/eventprocessing/pipelines/bulk", userAuth, platform, write)
	bp.POST("/create", bh.Create)
	bp.POST("/update", bh.Update)
	bp.POST("/delete", bh.Delete)
	bp.POST("/activate", bh.Activate)

	bcr := handler.NewBulkCorrelationRuleHandler(m.GetCorrelationRuleUsecase(), tenantLister)
	bcg := api.Group("/platform/eventprocessing/correlation-rule/bulk", userAuth, platform, write)
	bcg.POST("/create", bcr.Create)
	bcg.POST("/update", bcr.Update)
	bcg.POST("/delete", bcr.Delete)
	bcg.POST("/activate", bcr.Activate)
}
