package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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
// @Success     201 {object} domain.UtmIncidentAlert
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-incident-alerts [post]
func (h *IncidentAlertHandler) Create(c *gin.Context) {
	var req dto.IncidentAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	row, err := h.usecase.Create(c.Request.Context(), req)
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
// @Router      /utm-incident-alerts/update-status [post]
func (h *IncidentAlertHandler) UpdateStatus(c *gin.Context) {
	var req dto.UpdateAlertStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.usecase.UpdateStatus(c.Request.Context(), req); err != nil {
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
// @Success     200 {object} domain.UtmIncidentAlert
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-incident-alerts [put]
func (h *IncidentAlertHandler) Update(c *gin.Context) {
	var req dto.UpdateIncidentAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	row, err := h.usecase.Update(c.Request.Context(), req)
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
// @Param       incidentId   query int    false "Filter by incident ID"
// @Param       alertId      query string false "Filter by alert ID"
// @Param       alertStatus  query int    false "Filter by alert status"
// @Param       page         query int    false "Page (default 1)"
// @Param       size         query int    false "Page size (default 20)"
// @Success     200 {array} domain.UtmIncidentAlert
// @Header      200 {string} X-Total-Count "Total items"
// @Failure     500 {object} map[string]string
// @Router      /utm-incident-alerts [get]
func (h *IncidentAlertHandler) List(c *gin.Context) {
	query := dto.IncidentAlertListQuery{
		IncidentID:  queryInt64(c, "incidentId"),
		AlertID:     queryString(c, "alertId"),
		AlertStatus: queryIntPtr(c, "alertStatus"),
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
// @Param       id path int true "Incident alert ID"
// @Success     200 "Deleted"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-incident-alerts/{id} [delete]
func (h *IncidentAlertHandler) Delete(c *gin.Context) {
	id, ok := pathInt64(c, "id")
	if !ok {
		return
	}
	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		writeIncidentError(c, err)
		return
	}
	// Java sets X-utmstackApp-alert header on deletion.
	c.Header("X-utmstackApp-alert", "incidentAlert,"+strconv.FormatInt(id, 10))
	c.Status(http.StatusOK)
}
