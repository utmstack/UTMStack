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
	err := h.usecase.UpdateStatus(c.Request.Context(), c.GetString("user_email"), req)
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

	err = h.usecase.UpdateNotes(c.Request.Context(), c.GetString("user_email"), alertID, notes)
	audit.Record(c, audit_connectors.Event{Action: "alert.notes"}, audit_domain.ALERT_NOTE_UPDATE_ATTEMPT, audit_domain.ALERT_NOTE_UPDATE_SUCCESS, err)
	if err != nil {
		writeAlertError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (h *AlertHandler) UpdateAssignee(c *gin.Context) {
	var req dto.UpdateAlertAssigneeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.usecase.UpdateAssignee(c.Request.Context(), c.GetString("user_email"), req.AlertID, req.Assignee)
	audit.Record(c, audit_connectors.Event{Action: "alert.assignee"}, audit_domain.ALERT_UPDATE_ATTEMPT, audit_domain.ALERT_UPDATE_SUCCESS, err)
	if err != nil {
		writeAlertError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// @Summary     Related logs for an alert
// @Description Reproduces the Event Processor's correlation search (without the
// @Description engine's 10-hit cap) and returns the matching log ids so the UI can
// @Description load every related log in the Log Explorer.
// @Tags        Alerts
// @Security    BearerAuth
// @Produce     json
// @Param       alertId query string true "Alert ID"
// @Success     200 {object} dto.RelatedLogsResponse
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-alerts/related-logs [get]
func (h *AlertHandler) RelatedLogs(c *gin.Context) {
	alertID := c.Query("alertId")
	if alertID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing alert id"})
		return
	}
	resp, err := h.usecase.RelatedLogs(c.Request.Context(), alertID)
	if err != nil {
		writeAlertError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
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
	err := h.usecase.UpdateTags(c.Request.Context(), c.GetString("user_email"), req)
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
	err := h.usecase.ConvertToIncident(c.Request.Context(), c.GetString("user_email"), req)
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

// @Summary     List echoes of an alert
// @Description Returns the child alerts (parentId == :id) of a parent alert, paginated and sorted.
// @Tags        Alerts
// @Security    BearerAuth
// @Produce     json
// @Param       id        path  string true  "Parent alert id"
// @Param       page      query int    false "Page number (1-based, default 1)"
// @Param       size      query int    false "Page size (default 20, max 100)"
// @Param       sortBy    query string false "Sort field (default @timestamp)"
// @Param       sortOrder query string false "Sort order: asc|desc (default desc)"
// @Success     200 {array} domain.UtmAlert
// @Header      200 {string} X-Total-Count "Total matching echoes"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-alerts/{id}/echoes [get]
func (h *AlertHandler) ListEchoes(c *gin.Context) {
	parentID := c.Param("id")
	if parentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing alert id"})
		return
	}
	page := queryInt(c, "page", 1)
	size := queryInt(c, "size", 20)
	sortBy := c.Query("sortBy")
	sortOrder := c.DefaultQuery("sortOrder", "desc")

	items, total, err := h.usecase.ListEchoes(c.Request.Context(), parentID, page, size, sortBy, sortOrder)
	if err != nil {
		writeAlertError(c, err)
		return
	}
	writePagedArray(c, items, total)
}
