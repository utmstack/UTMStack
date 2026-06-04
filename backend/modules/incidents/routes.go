package incidents

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	ih := m.GetIncidentHandler()
	iah := m.GetIncidentAlertHandler()
	inh := m.GetIncidentNoteHandler()
	ihh := m.GetIncidentHistoryHandler()

	// Incidents
	ig := api.Group("/utm-incidents", userAuth)
	ig.POST("", ih.Create)
	ig.POST("/add-alerts", ih.AddAlerts)
	ig.PUT("/change-status", ih.ChangeStatus)
	ig.GET("", ih.List)
	ig.GET("/users-assigned", ih.GetUsersAssigned)
	ig.GET("/:id", ih.GetByID)

	// Incident alerts
	ag := api.Group("/utm-incident-alerts", userAuth)
	ag.POST("", iah.Create)
	ag.POST("/update-status", iah.UpdateStatus)
	ag.PUT("", iah.Update)
	ag.GET("", iah.List)
	ag.DELETE("/:id", iah.Delete)

	// Incident notes
	ng := api.Group("/utm-incident-notes", userAuth)
	ng.POST("", inh.Create)
	ng.GET("", inh.List)

	// Incident histories
	hg := api.Group("/utm-incident-histories", userAuth)
	hg.GET("", ihh.List)
	hg.GET("/count", ihh.Count)
	hg.GET("/:id", ihh.GetByID)
}
