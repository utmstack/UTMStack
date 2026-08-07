package incidents

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	ih := m.GetIncidentHandler()
	iah := m.GetIncidentAlertHandler()
	inh := m.GetIncidentNoteHandler()
	ihh := m.GetIncidentHistoryHandler()

	read := middleware.RequirePermission("incidents.read")
	write := middleware.RequirePermission("incidents.write")

	// Incidents
	ig := api.Group("/incidents", userAuth)
	ig.POST("", write, ih.Create)
	ig.POST("/add-alerts", write, ih.AddAlerts)
	ig.PUT("/change-status", write, ih.ChangeStatus)
	ig.PUT("/assign", write, ih.Assign)
	ig.GET("", read, ih.List)
	ig.GET("/assignees", read, ih.GetAssignees)
	ig.GET("/:id", read, ih.GetByID)

	// Incident alerts
	ag := api.Group("/incident-alerts", userAuth)
	ag.POST("", write, iah.Create)
	ag.POST("/update-status", write, iah.UpdateStatus)
	ag.PUT("", write, iah.Update)
	ag.GET("", read, iah.List)
	ag.DELETE("/:id", write, iah.Delete)

	// Incident notes
	ng := api.Group("/incident-notes", userAuth)
	ng.POST("", write, inh.Create)
	ng.GET("", read, inh.List)

	// Incident histories
	hg := api.Group("/incident-histories", userAuth)
	hg.GET("", read, ihh.List)
	hg.GET("/count", read, ihh.Count)
	hg.GET("/:id", read, ihh.GetByID)
}
