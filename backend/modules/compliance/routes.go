package compliance

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	read := middleware.RequirePermission("compliance.read")
	write := middleware.RequirePermission("compliance.write")

	fw := api.Group("/compliance", userAuth)
	fw.GET("/controls", read, m.frameworkH.ListControls)
	fw.GET("/controls/:id", read, m.frameworkH.GetControl)
	fw.POST("/controls", write, m.frameworkH.CreateControl)
	fw.PUT("/controls/:id", write, m.frameworkH.UpdateControl)
	fw.DELETE("/controls/:id", write, m.frameworkH.DeleteControl)
	fw.PUT("/controls/:id/enabled", write, m.frameworkH.SetControlEnabled)

	fw.GET("/frameworks", read, m.frameworkH.ListFrameworks)
	fw.GET("/frameworks/:key", read, m.frameworkH.GetFramework)
	fw.GET("/frameworks/:key/report", read, m.reportH.GetFrameworkReport)        // evaluate now → report JSON (live)
	fw.GET("/frameworks/:key/report.pdf", read, m.reportH.GetFrameworkReportPDF) // live eval → PDF download
	fw.POST("/frameworks/:key/report", write, m.reportH.GenerateReport)          // evaluate + store snapshot
	// Stored report snapshots (history for the frontend).
	fw.GET("/reports", read, m.reportH.ListReports)
	fw.GET("/reports/:id", read, m.reportH.GetReport)
	fw.DELETE("/reports/:id", write, m.reportH.DeleteReport)
	fw.GET("/reports/:id/pdf", read, m.reportH.GetSnapshotPDF) // snapshot → PDF download
	fw.POST("/frameworks", write, m.frameworkH.CreateFramework)
	fw.PUT("/frameworks/:key", write, m.frameworkH.UpdateFramework)
	fw.DELETE("/frameworks/:key", write, m.frameworkH.DeleteFramework)
	fw.PUT("/frameworks/:key/enabled", write, m.frameworkH.SetFrameworkEnabled)

	fw.PUT("/frameworks/:key/controls/:id/status", write, m.reportH.SetStatusOverride)
	fw.DELETE("/frameworks/:key/controls/:id/status", write, m.reportH.ClearStatusOverride)
	fw.PUT("/frameworks/:key/controls/:id/note", write, m.reportH.SetControlNote)
	fw.DELETE("/frameworks/:key/controls/:id/note", write, m.reportH.ClearControlNote)

	sched := api.Group("/compliance-report-schedules", userAuth)
	sched.POST("", write, m.scheduleH.Create)
	sched.PUT("", write, m.scheduleH.Update)
	sched.GET("/by-user", read, m.scheduleH.ListByUser)
	sched.GET("/by-id/:id", read, m.scheduleH.GetByID)
	sched.DELETE("/:id", write, m.scheduleH.Delete)
}
