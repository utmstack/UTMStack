package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/incidents/connectors"
	"github.com/utmstack/utmstack/backend/modules/incidents/dto"
)

type IncidentNoteHandler struct {
	usecase connectors.IncidentNoteUsecase
}

func NewIncidentNoteHandler(uc connectors.IncidentNoteUsecase) *IncidentNoteHandler {
	return &IncidentNoteHandler{usecase: uc}
}

// @Summary     Create incident note
// @Tags        Incidents
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.CreateNoteRequest true "Create note"
// @Success     201 {object} domain.UtmIncidentNote
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-incident-notes [post]
func (h *IncidentNoteHandler) Create(c *gin.Context) {
	var req dto.CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	note, err := h.usecase.Create(c.Request.Context(), req)
	if err != nil {
		writeIncidentError(c, err)
		return
	}
	c.JSON(http.StatusCreated, note)
}

// @Summary     List incident notes
// @Tags        Incidents
// @Security    BearerAuth
// @Produce     json
// @Param       incidentId query int false "Filter by incident ID"
// @Param       page       query int false "Page (default 1)"
// @Param       size       query int false "Page size (default 20)"
// @Success     200 {array} domain.UtmIncidentNote
// @Header      200 {string} X-Total-Count "Total items"
// @Failure     500 {object} map[string]string
// @Router      /utm-incident-notes [get]
func (h *IncidentNoteHandler) List(c *gin.Context) {
	query := dto.IncidentNoteListQuery{
		IncidentID: queryInt64(c, "incidentId"),
		Page:       queryInt(c, "page", 1),
		Size:       queryInt(c, "size", 20),
	}
	rows, total, err := h.usecase.List(c.Request.Context(), query)
	if err != nil {
		writeIncidentError(c, err)
		return
	}
	writePagedArray(c, rows, total)
}
