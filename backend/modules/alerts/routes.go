package alerts

import (
	"github.com/gin-gonic/gin"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	auditmw "github.com/utmstack/utmstack/backend/modules/audit/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc, audit audit_connectors.Logger) {
	h := m.GetAlertHandler()
	th := m.GetAlertTagHandler()
	rh := m.GetAlertTagRuleHandler()

	audited := func(
		attemptType audit_domain.ApplicationEventType, attemptMsg string,
		successType audit_domain.ApplicationEventType, successMsg string,
	) gin.HandlerFunc {
		return auditmw.AuditEvent(audit, attemptType, attemptMsg, successType, successMsg)
	}

	ag := api.Group("/utm-alerts", userAuth)

	ag.POST("/status",
		audited(
			audit_domain.ALERT_UPDATE_ATTEMPT, "alert.status.attempt",
			audit_domain.ALERT_UPDATE_SUCCESS, "alert.status.success",
		),
		h.UpdateStatus,
	)

	ag.POST("/notes",
		audited(
			audit_domain.ALERT_NOTE_UPDATE_ATTEMPT, "alert.notes.attempt",
			audit_domain.ALERT_NOTE_UPDATE_SUCCESS, "alert.notes.success",
		),
		h.UpdateNotes,
	)

	ag.POST("/tags",
		audited(
			audit_domain.ALERT_TAG_UPDATE_ATTEMPT, "alert.tags.attempt",
			audit_domain.ALERT_TAG_UPDATE_SUCCESS, "alert.tags.success",
		),
		h.UpdateTags,
	)

	ag.POST("/convert-to-incident",
		audited(
			audit_domain.ALERT_CONVERT_TO_INCIDENT_ATTEMPT, "alert.convert_to_incident.attempt",
			audit_domain.ALERT_CONVERT_TO_INCIDENT_SUCCESS, "alert.convert_to_incident.success",
		),
		h.ConvertToIncident,
	)

	ag.GET("/count-open-alerts", h.CountOpenAlerts)

	tg := api.Group("/utm-alert-tags", userAuth)
	tg.POST("", th.Create)
	tg.PUT("", th.Update)
	tg.GET("", th.List)
	tg.GET("/:id", th.GetByID)
	tg.DELETE("/:id", th.Delete)

	rg := api.Group("/alert-tag-rules", userAuth)
	rg.POST("", rh.Create)
	rg.PUT("", rh.Update)
	rg.GET("", rh.List)
	rg.GET("/get-by-ids", rh.GetByIDs)
	rg.GET("/:id", rh.GetByID)
	rg.DELETE("/:id", rh.Delete)
}
