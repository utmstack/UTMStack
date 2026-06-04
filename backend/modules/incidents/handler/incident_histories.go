package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/incidents/connectors"
	"github.com/utmstack/utmstack/backend/modules/incidents/dto"
)

type IncidentHistoryHandler struct {
	usecase connectors.IncidentHistoryUsecase
}

func NewIncidentHistoryHandler(uc connectors.IncidentHistoryUsecase) *IncidentHistoryHandler {
	return &IncidentHistoryHandler{usecase: uc}
}

// @Summary     List incident histories
// @Tags        Incidents
// @Security    BearerAuth
// @Produce     json
// @Param       incidentId query int false "Filter by incident ID"
// @Param       page       query int false "Page (default 1)"
// @Param       size       query int false "Page size (default 20)"
// @Param       sort       query string false "Sort field (default action_date DESC)"
// @Success     200 {array} domain.UtmIncidentHistory
// @Header      200 {string} X-Total-Count "Total items"
// @Failure     500 {object} map[string]string
// @Router      /utm-incident-histories [get]
func (h *IncidentHistoryHandler) List(c *gin.Context) {
	query := dto.HistoryListQuery{
		IncidentID: queryInt64(c, "incidentId"),
		ActionType: queryString(c, "actionType"),
		Page:       queryInt(c, "page", 1),
		Size:       queryInt(c, "size", 20),
		Sort:       c.Query("sort"),
	}
	rows, total, err := h.usecase.List(c.Request.Context(), query)
	if err != nil {
		writeIncidentError(c, err)
		return
	}
	writePagedArray(c, rows, total)
}

// @Summary     Count incident histories
// @Tags        Incidents
// @Security    BearerAuth
// @Produce     json
// @Param       incidentId query int false "Filter by incident ID"
// @Success     200 {integer} int64
// @Failure     500 {object} map[string]string
// @Router      /utm-incident-histories/count [get]
func (h *IncidentHistoryHandler) Count(c *gin.Context) {
	query := dto.HistoryListQuery{
		IncidentID: queryInt64(c, "incidentId"),
	}
	total, err := h.usecase.Count(c.Request.Context(), query)
	if err != nil {
		writeIncidentError(c, err)
		return
	}
	c.JSON(http.StatusOK, total)
}

// @Summary     Get incident history by ID
// @Tags        Incidents
// @Security    BearerAuth
// @Produce     json
// @Param       id path int true "History ID"
// @Success     200 {object} domain.UtmIncidentHistory
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-incident-histories/{id} [get]
func (h *IncidentHistoryHandler) GetByID(c *gin.Context) {
	id, ok := pathInt64(c, "id")
	if !ok {
		return
	}
	h2, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		writeIncidentError(c, err)
		return
	}
	c.JSON(http.StatusOK, h2)
}
