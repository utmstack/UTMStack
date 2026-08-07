package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/incidents/connectors"
	"github.com/utmstack/utmstack/backend/modules/incidents/dto"
)

type IncidentAlertHandler struct {
	usecase connectors.IncidentAlertUsecase
}

func NewIncidentAlertHandler(uc connectors.IncidentAlertUsecase) *IncidentAlertHandler {
	return &IncidentAlertHandler{usecase: uc}
}

// @Summary     Create incident alert
// @Tags        Incidents
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.IncidentAlertRequest true "Incident alert"
// @Success     201 {object} domain.IncidentAlert
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /incident-alerts [post]
func (h *IncidentAlertHandler) Create(c *gin.Context) {
	var req dto.IncidentAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	row, err := h.usecase.Create(c.Request.Context(), req)
	audit.Record(c, audit_connectors.Event{Action: "incident.alert.create", ResourceType: "incident_alert"},
		audit_domain.INCIDENT_ALERT_ADD_ATTEMPT, audit_domain.INCIDENT_ALERT_ADD_SUCCESS, err)
	if err != nil {
		writeIncidentError(c, err)
		return
	}
	c.JSON(http.StatusCreated, row)
}

// @Summary     Update incident alert statuses
// @Tags        Incidents
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.UpdateAlertStatusRequest true "Status update"
// @Success     200 "Updated"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /incident-alerts/update-status [post]
func (h *IncidentAlertHandler) UpdateStatus(c *gin.Context) {
	var req dto.UpdateAlertStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.usecase.UpdateStatus(c.Request.Context(), c.GetString("user_email"), req)
	audit.Record(c, audit_connectors.Event{Action: "incident.alert.update_status", ResourceType: "incident_alert"},
		audit_domain.INCIDENT_UPDATE_ATTEMPT, audit_domain.INCIDENT_UPDATE_SUCCESS, err)
	if err != nil {
		writeIncidentError(c, err)
		return
	}
	c.Status(http.StatusOK)
}

// @Summary     Update incident alert
// @Tags        Incidents
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.UpdateIncidentAlertRequest true "Incident alert update"
// @Success     200 {object} domain.IncidentAlert
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /incident-alerts [put]
func (h *IncidentAlertHandler) Update(c *gin.Context) {
	var req dto.UpdateIncidentAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	row, err := h.usecase.Update(c.Request.Context(), req)
	audit.Record(c, audit_connectors.Event{Action: "incident.alert.update", ResourceType: "incident_alert"},
		audit_domain.INCIDENT_UPDATE_ATTEMPT, audit_domain.INCIDENT_UPDATE_SUCCESS, err)
	if err != nil {
		writeIncidentError(c, err)
		return
	}
	c.JSON(http.StatusOK, row)
}

// @Summary     List incident alerts
// @Tags        Incidents
// @Security    BearerAuth
// @Produce     json
// @Param       incidentId   query string false "Filter by incident ID"
// @Param       alertId      query string false "Filter by alert ID"
// @Param       alertStatus  query string false "Filter by alert status"
// @Param       page         query int    false "Page (default 1)"
// @Param       size         query int    false "Page size (default 20)"
// @Success     200 {array} domain.IncidentAlert
// @Header      200 {string} X-Total-Count "Total items"
// @Failure     500 {object} map[string]string
// @Router      /incident-alerts [get]
func (h *IncidentAlertHandler) List(c *gin.Context) {
	incidentID, ok := queryUUID(c, "incidentId")
	if !ok {
		return
	}
	query := dto.IncidentAlertListQuery{
		IncidentID:  incidentID,
		AlertID:     queryString(c, "alertId"),
		AlertStatus: queryString(c, "alertStatus"),
		Page:        queryInt(c, "page", 1),
		Size:        queryInt(c, "size", 20),
	}
	rows, total, err := h.usecase.List(c.Request.Context(), query)
	if err != nil {
		writeIncidentError(c, err)
		return
	}
	writePagedArray(c, rows, total)
}

// @Summary     Delete incident alert
// @Tags        Incidents
// @Security    BearerAuth
// @Param       id path string true "Incident alert ID"
// @Success     200 "Deleted"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /incident-alerts/{id} [delete]
func (h *IncidentAlertHandler) Delete(c *gin.Context) {
	id, ok := pathUUID(c, "id")
	if !ok {
		return
	}
	err := h.usecase.Delete(c.Request.Context(), id)
	audit.Record(c, audit_connectors.Event{Action: "incident.alert.delete", ResourceType: "incident_alert", ResourceID: id.String()},
		audit_domain.INCIDENT_UPDATE_ATTEMPT, audit_domain.INCIDENT_UPDATE_SUCCESS, err)
	if err != nil {
		writeIncidentError(c, err)
		return
	}
	// Java sets X-utmstackApp-alert header on deletion.
	c.Header("X-utmstackApp-alert", "incidentAlert,"+id.String())
	c.Status(http.StatusOK)
}
