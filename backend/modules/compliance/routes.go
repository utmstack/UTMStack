package compliance

import (
	"github.com/gin-gonic/gin"

	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc, platform gin.HandlerFunc) {
	read := middleware.RequirePermission("compliance.read")
	write := middleware.RequirePermission("compliance.write")

	fw := api.Group("/compliance", userAuth)

	fw.GET("/controls", read, m.frameworkH.ListControls)
	fw.GET("/controls/:id", read, m.frameworkH.GetControl)
	fw.POST("/controls", write, m.frameworkH.CreateControl)
	fw.PUT("/controls/:id", write, m.frameworkH.UpdateControl)
	fw.DELETE("/controls/:id", write, m.frameworkH.DeleteControl)

	fw.GET("/frameworks", read, m.frameworkH.ListFrameworks)
	fw.GET("/frameworks/:key", read, m.frameworkH.GetFramework)
	fw.POST("/frameworks", write, m.frameworkH.CreateFramework)
	fw.PUT("/frameworks/:key", write, m.frameworkH.UpdateFramework)
	fw.DELETE("/frameworks/:key", write, m.frameworkH.DeleteFramework)

	fw.GET("/reports", read, m.reportH.ListReports)
	fw.GET("/frameworks/:key/report", read, m.reportH.GetReport)
	fw.POST("/frameworks/:key/report", write, m.reportH.Evaluate)
	fw.DELETE("/frameworks/:key/report", write, m.reportH.DeleteReport)
	fw.GET("/frameworks/:key/report.pdf", read, m.reportH.GetReportPDF)

	fw.PUT("/frameworks/:key/controls/:id/status", write, m.reportH.EditControl)

	fw.GET("/frameworks/:key/history", read, m.reportH.History)
	fw.GET("/frameworks/:key/history.pdf", read, m.reportH.HistoryPDF)

	sched := api.Group("/compliance-report-schedules", userAuth)
	sched.POST("", write, m.scheduleH.Create)
	sched.PUT("", write, m.scheduleH.Update)
	sched.GET("/by-user", read, m.scheduleH.ListByUser)
	sched.GET("/by-id/:id", read, m.scheduleH.GetByID)
	sched.DELETE("/:id", write, m.scheduleH.Delete)

	// Platform-admin bulk operations (cross-tenant).
	bh := m.GetBulkHandler()
	pg := api.Group("/platform/compliance", userAuth, platform, write)
	pg.POST("/frameworks/bulk/create", bh.CreateFramework)
	pg.POST("/frameworks/bulk/update", bh.UpdateFramework)
	pg.POST("/frameworks/bulk/delete", bh.DeleteFramework)
	pg.POST("/controls/bulk/create", bh.CreateControl)
	pg.POST("/controls/bulk/update", bh.UpdateControl)
	pg.POST("/controls/bulk/delete", bh.DeleteControl)
}
