package alerts

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	h := m.GetAlertHandler()
	th := m.GetAlertTagHandler()
	rh := m.GetAlertTagRuleHandler()
	dh := m.GetAdversaryHandler()

	read := middleware.RequirePermission("alerts.read")
	write := middleware.RequirePermission("alerts.write")

	ag := api.Group("/utm-alerts", userAuth)

	ag.POST("/status", write, h.UpdateStatus)
	ag.POST("/notes", write, h.UpdateNotes)
	ag.POST("/assignee", write, h.UpdateAssignee)
	ag.POST("/tags", write, h.UpdateTags)
	ag.POST("/convert-to-incident", write, h.ConvertToIncident)

	ag.GET("/count-open-alerts", read, h.CountOpenAlerts)
	ag.GET("/related-logs", read, h.RelatedLogs)
	ag.GET("/:id/echoes", read, h.ListEchoes)

	tg := api.Group("/utm-alert-tags", userAuth)
	tg.POST("", write, th.Create)
	tg.PUT("", write, th.Update)
	tg.GET("", read, th.List)
	tg.GET("/:id", read, th.GetByID)
	tg.DELETE("/:id", write, th.Delete)

	rg := api.Group("/alert-tag-rules", userAuth)
	rg.POST("", write, rh.Create)
	rg.PUT("", write, rh.Update)
	rg.GET("", read, rh.List)
	rg.GET("/get-by-ids", read, rh.GetByIDs)
	rg.GET("/:id", read, rh.GetByID)
	rg.DELETE("/:id", write, rh.Delete)

	// Internal-only: alerts plugin polls this for the active rule set.
	ig := api.Group("/internal/alert-tag-rules", userAuth, middleware.RequireInternal())
	ig.GET("/active", rh.ListActive)

	ia := api.Group("/internal/alerts", userAuth, middleware.RequireInternal())
	ia.POST("/notify", h.NotifyRaised)

	// Adversary alert aggregation (merged from the former threat_management module).
	adv := api.Group("/adversary", userAuth)
	adv.POST("/alerts", read, dh.SearchAdversaryAlerts)
}
