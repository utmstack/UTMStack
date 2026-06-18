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

type IncidentHandler struct {
	usecase connectors.IncidentUsecase
}

func NewIncidentHandler(uc connectors.IncidentUsecase) *IncidentHandler {
	return &IncidentHandler{usecase: uc}
}

// @Summary     Create incident
// @Tags        Incidents
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.CreateIncidentRequest true "Create incident"
// @Success     200 {object} domain.UtmIncident
// @Failure     400 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /incidents [post]
func (h *IncidentHandler) Create(c *gin.Context) {
	var req dto.CreateIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	incident, err := h.usecase.Create(c.Request.Context(), c.GetString("user_login"), req)
	audit.Record(c, audit_connectors.Event{Action: "incident.create", ResourceType: "incident"},
		audit_domain.INCIDENT_CREATION_ATTEMPT, audit_domain.INCIDENT_CREATION_SUCCESS, err)
	if err != nil {
		writeIncidentError(c, err)
		return
	}
	c.Header("X-utmstackApp-alert", "incident,"+req.IncidentName)
	c.JSON(http.StatusOK, incident)
}

// @Summary     Add alerts to incident
// @Tags        Incidents
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.AddAlertsRequest true "Add alerts"
// @Success     201 {object} domain.UtmIncident
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /incidents/add-alerts [post]
func (h *IncidentHandler) AddAlerts(c *gin.Context) {
	var req dto.AddAlertsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	incident, err := h.usecase.AddAlerts(c.Request.Context(), c.GetString("user_login"), req)
	audit.Record(c, audit_connectors.Event{Action: "incident.add_alerts", ResourceType: "incident"},
		audit_domain.INCIDENT_ALERT_ADD_ATTEMPT, audit_domain.INCIDENT_ALERT_ADD_SUCCESS, err)
	if err != nil {
		writeIncidentError(c, err)
		return
	}
	// Java returns 201 Created for POST /incidents/add-alerts.
	c.JSON(http.StatusCreated, incident)
}

// @Summary     Change incident status
// @Tags        Incidents
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.ChangeStatusRequest true "Change status"
// @Success     200 {object} domain.UtmIncident
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /incidents/change-status [put]
func (h *IncidentHandler) ChangeStatus(c *gin.Context) {
	var req dto.ChangeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	incident, err := h.usecase.ChangeStatus(c.Request.Context(), c.GetString("user_login"), req)
	audit.Record(c, audit_connectors.Event{Action: "incident.change_status", ResourceType: "incident"},
		audit_domain.INCIDENT_UPDATE_ATTEMPT, audit_domain.INCIDENT_UPDATE_SUCCESS, err)
	if err != nil {
		writeIncidentError(c, err)
		return
	}
	c.JSON(http.StatusOK, incident)
}

// @Summary     Assign incident
// @Description Assigns or (with a null/empty assignedTo) unassigns an incident.
// @Tags        Incidents
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.AssignRequest true "Assignment"
// @Success     200 {object} domain.UtmIncident
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Router      /incidents/assign [put]
func (h *IncidentHandler) Assign(c *gin.Context) {
	var req dto.AssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	incident, err := h.usecase.Assign(c.Request.Context(), c.GetString("user_login"), req)
	audit.Record(c, audit_connectors.Event{Action: "incident.assign", ResourceType: "incident"},
		audit_domain.INCIDENT_UPDATE_ATTEMPT, audit_domain.INCIDENT_UPDATE_SUCCESS, err)
	if err != nil {
		writeIncidentError(c, err)
		return
	}
	c.JSON(http.StatusOK, incident)
}

// @Summary     List incidents
// @Tags        Incidents
// @Security    BearerAuth
// @Produce     json
// @Param       incidentName       query string false "Filter by name (ILIKE)"
// @Param       incidentStatus     query string false "Filter by status"
// @Param       incidentAssignedTo query string false "Filter by assigned user"
// @Param       page               query int    false "Page (default 1)"
// @Param       size               query int    false "Page size (default 20)"
// @Success     200 {array} domain.UtmIncident
// @Header      200 {string} X-Total-Count "Total items"
// @Failure     500 {object} map[string]string
// @Router      /incidents [get]
func (h *IncidentHandler) List(c *gin.Context) {
	query := dto.IncidentListQuery{
		IncidentName:       queryString(c, "incidentName"),
		IncidentStatus:     queryString(c, "incidentStatus"),
		IncidentAssignedTo: queryString(c, "incidentAssignedTo"),
		CreatedDateStart:   queryTime(c, "createdDateStart"),
		CreatedDateEnd:     queryTime(c, "createdDateEnd"),
		Page:               queryInt(c, "page", 1),
		Size:               queryInt(c, "size", 20),
		Sort:               c.Query("sort"),
	}
	rows, total, err := h.usecase.List(c.Request.Context(), query)
	if err != nil {
		writeIncidentError(c, err)
		return
	}
	writePagedArray(c, rows, total)
}

// @Summary     Get users assigned to incidents
// @Tags        Incidents
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} dto.UserAssignedDTO
// @Failure     500 {object} map[string]string
// @Router      /incidents/users-assigned [get]
func (h *IncidentHandler) GetUsersAssigned(c *gin.Context) {
	users, err := h.usecase.GetUsersAssigned(c.Request.Context())
	if err != nil {
		writeIncidentError(c, err)
		return
	}
	if users == nil {
		users = []dto.UserAssignedDTO{}
	}
	c.JSON(http.StatusOK, users)
}

// @Summary     Get incident by ID
// @Tags        Incidents
// @Security    BearerAuth
// @Produce     json
// @Param       id path int true "Incident ID"
// @Success     200 {object} domain.UtmIncident
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /incidents/{id} [get]
func (h *IncidentHandler) GetByID(c *gin.Context) {
	id, ok := pathInt64(c, "id")
	if !ok {
		return
	}
	incident, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		writeIncidentError(c, err)
		return
	}
	c.JSON(http.StatusOK, incident)
}
