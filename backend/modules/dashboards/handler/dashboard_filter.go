package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/dashboards/connectors"
	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/modules/dashboards/dto"
)

type DashboardFilterHandler struct {
	uc connectors.DashboardFilterUsecase
}

func NewDashboardFilterHandler(uc connectors.DashboardFilterUsecase) *DashboardFilterHandler {
	return &DashboardFilterHandler{uc: uc}
}

// Create godoc
//
//	@Summary		Create a dashbaord filter
//	@Description	Creates a custom dashboard filter.
//	@Tags			DashboardFilters
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		domain.DashboardFilter	true	"DashboardFilter to create"
//	@Success		201		{object}	domain.DashboardFilter
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/filter [post]
func (h *DashboardFilterHandler) Create(c *gin.Context) {
	var v domain.DashboardFilter
	if err := c.ShouldBindJSON(&v); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.uc.Create(c.Request.Context(), &v, currentUser(c))
	resID := ""
	if res != nil {
		resID = res.ID.String()
	}
	audit.Record(c, audit_connectors.Event{Action: "dashboard_filter.create", ResourceType: "dashboard_filter", ResourceID: resID},
		audit_domain.DASHBOARD_FILTER_CREATE_ATTEMPT, audit_domain.DASHBOARD_FILTER_CREATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, res)
}

// Update godoc
//
//	@Summary		Update a dashbaord filter
//	@Description	Updates an existing dashboard filter
//	@Tags			DashboardFilters
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		domain.DashboardFilter	true	"DashboardFilter to update"
//	@Success		200		{object}	domain.DashboardFilter
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/filter [put]
func (h *DashboardFilterHandler) Update(c *gin.Context) {
	var v domain.DashboardFilter
	if err := c.ShouldBindJSON(&v); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.uc.Update(c.Request.Context(), &v, currentUser(c))
	audit.Record(c, audit_connectors.Event{Action: "dashboard_filter.update", ResourceType: "dashboard_filter", ResourceID: v.ID.String()},
		audit_domain.DASHBOARD_FILTER_UPDATE_ATTEMPT, audit_domain.DASHBOARD_FILTER_UPDATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// List godoc
//
//	@Summary		List dashbaord filters
//	@Description	Lists dashboard filter.
//	@Tags			DashboardFilters
//	@Security		BearerAuth
//	@Produce		json
//	@Param			dashboardId	query		int		false	"Filter by dashboard id"
//	@Param			page		query		int		false	"Page (0-based)"
//	@Param			size		query		int		false	"Page size"
//	@Success		200			{array}		domain.DashboardFilter
//	@Header			200			{string}	X-Total-Count	"Total records"
//	@Failure		500			{object}	map[string]string
//	@Router			/filters [get]
func (h *DashboardFilterHandler) List(c *gin.Context) {
	var q dto.DashboardFilterQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	f, err := q.Filter()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dashboardId must be a uuid"})
		return
	}
	items, total, err := h.uc.List(c.Request.Context(), f)
	if err != nil {
		writeError(c, err)
		return
	}
	writeList(c, items, total)
}


// Delete godoc
//
//	@Summary		Delete a dashbaord filter by id
//	@Tags			DashboardFilters
//	@Security		BearerAuth
//	@Param			id	path	int	true	"DashboardFilter id"
//	@Success		200	"Deleted"
//	@Failure		500	{object}	map[string]string
//	@Router			/filter/{id} [delete]
func (h *DashboardFilterHandler) Delete(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	err := h.uc.Delete(c.Request.Context(), id)
	audit.Record(c, audit_connectors.Event{Action: "visualization.delete", ResourceType: "visualization", ResourceID: id.String()},
		audit_domain.VISUALIZATION_DELETE_ATTEMPT, audit_domain.VISUALIZATION_DELETE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusOK)
}
