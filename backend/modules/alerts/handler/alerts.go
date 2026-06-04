package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/modules/alerts/dto"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
)

type AlertHandler struct {
	usecase connectors.AlertUsecase
}

func NewAlertHandler(uc connectors.AlertUsecase) *AlertHandler {
	return &AlertHandler{usecase: uc}
}

// @Summary     Update alert status
// @Tags        Alerts
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.UpdateAlertStatusRequest true "Status update"
// @Success     200 "Updated"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-alerts/status [post]
func (h *AlertHandler) UpdateStatus(c *gin.Context) {
	var req dto.UpdateAlertStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.usecase.UpdateStatus(c.Request.Context(), c.GetString("user_login"), req)
	audit.Record(c, audit_connectors.Event{Action: "alert.status"}, audit_domain.ALERT_UPDATE_ATTEMPT, audit_domain.ALERT_UPDATE_SUCCESS, err)
	if err != nil {
		writeAlertError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// @Summary     Update alert notes
// @Tags        Alerts
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       alertId query  string true "Alert ID"
// @Param       body    body   string true "Notes content"
// @Success     200 "Updated"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-alerts/notes [post]
func (h *AlertHandler) UpdateNotes(c *gin.Context) {
	alertID := c.Query("alertId")
	if alertID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing alert id"})
		return
	}

	bodyBytes, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not read request body"})
		return
	}
	notes := string(bodyBytes)

	err = h.usecase.UpdateNotes(c.Request.Context(), c.GetString("user_login"), alertID, notes)
	audit.Record(c, audit_connectors.Event{Action: "alert.notes"}, audit_domain.ALERT_NOTE_UPDATE_ATTEMPT, audit_domain.ALERT_NOTE_UPDATE_SUCCESS, err)
	if err != nil {
		writeAlertError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// @Summary     Update alert tags
// @Tags        Alerts
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.UpdateAlertTagsRequest true "Tags update"
// @Success     200 "Updated"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-alerts/tags [post]
func (h *AlertHandler) UpdateTags(c *gin.Context) {
	var req dto.UpdateAlertTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.usecase.UpdateTags(c.Request.Context(), c.GetString("user_login"), req)
	audit.Record(c, audit_connectors.Event{Action: "alert.tags"}, audit_domain.ALERT_TAG_UPDATE_ATTEMPT, audit_domain.ALERT_TAG_UPDATE_SUCCESS, err)
	if err != nil {
		writeAlertError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// @Summary     Convert alerts to incident
// @Tags        Alerts
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.ConvertToIncidentRequest true "Convert request"
// @Success     200 "Converted"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-alerts/convert-to-incident [post]
func (h *AlertHandler) ConvertToIncident(c *gin.Context) {
	var req dto.ConvertToIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.usecase.ConvertToIncident(c.Request.Context(), c.GetString("user_login"), req)
	audit.Record(c, audit_connectors.Event{Action: "alert.convert_to_incident"}, audit_domain.ALERT_CONVERT_TO_INCIDENT_ATTEMPT, audit_domain.ALERT_CONVERT_TO_INCIDENT_SUCCESS, err)
	if err != nil {
		writeAlertError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// @Summary     Count open alerts
// @Tags        Alerts
// @Security    BearerAuth
// @Produce     json
// @Success     200 {integer} int64
// @Failure     500 {object} map[string]string
// @Router      /utm-alerts/count-open-alerts [get]
func (h *AlertHandler) CountOpenAlerts(c *gin.Context) {
	resp, err := h.usecase.CountOpenAlerts(c.Request.Context())
	if err != nil {
		writeAlertError(c, err)
		return
	}
	// Java returns bare Long — match that contract.
	c.JSON(http.StatusOK, resp.Count)
}
