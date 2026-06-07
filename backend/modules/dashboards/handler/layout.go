package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/dashboards/connectors"
	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/modules/dashboards/dto"
)

type LayoutHandler struct{ uc connectors.LayoutUsecase }

func NewLayoutHandler(uc connectors.LayoutUsecase) *LayoutHandler {
	return &LayoutHandler{uc: uc}
}

// Create godoc
//
//	@Summary		Add a visualization to a dashboard (layout placement)
//	@Description	Places a visualization on a dashboard with its grid layout (order/size/position). The id must be absent.
//	@Tags			Dashboard Layouts
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		domain.UtmDashboardVisualization	true	"Layout placement to create"
//	@Success		201		{object}	domain.UtmDashboardVisualization
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/dashboard-layouts [post]
func (h *LayoutHandler) Create(c *gin.Context) {
	var l domain.UtmDashboardVisualization
	if err := c.ShouldBindJSON(&l); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.uc.Create(c.Request.Context(), &l)
	audit.Record(c, audit_connectors.Event{Action: "dashboard.layout.create", ResourceType: "dashboard_visualization", ResourceID: strconv.FormatUint(l.IDDashboard, 10)},
		audit_domain.DASHBOARD_LAYOUT_CREATE_ATTEMPT, audit_domain.DASHBOARD_LAYOUT_CREATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, res)
}

// Update godoc
//
//	@Summary		Update a layout placement
//	@Description	Updates an existing dashboard↔visualization placement (order/size/position). The id is required.
//	@Tags			Dashboard Layouts
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		domain.UtmDashboardVisualization	true	"Layout placement to update"
//	@Success		200		{object}	domain.UtmDashboardVisualization
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/dashboard-layouts [put]
func (h *LayoutHandler) Update(c *gin.Context) {
	var l domain.UtmDashboardVisualization
	if err := c.ShouldBindJSON(&l); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.uc.Update(c.Request.Context(), &l)
	audit.Record(c, audit_connectors.Event{Action: "dashboard.layout.update", ResourceType: "dashboard_visualization", ResourceID: strconv.FormatUint(l.ID, 10)},
		audit_domain.DASHBOARD_LAYOUT_UPDATE_ATTEMPT, audit_domain.DASHBOARD_LAYOUT_UPDATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// List godoc
//
//	@Summary		List layout placements
//	@Description	Lists dashboard↔visualization placements, typically filtered by dashboard.
//	@Tags			Dashboard Layouts
//	@Security		BearerAuth
//	@Produce		json
//	@Param			idDashboard		query		int	false	"Filter by dashboard id"
//	@Param			idVisualization	query		int	false	"Filter by visualization id"
//	@Param			page			query		int	false	"Page (0-based)"
//	@Param			size			query		int	false	"Page size"
//	@Success		200				{array}		domain.UtmDashboardVisualization
//	@Header			200				{string}	X-Total-Count	"Total records"
//	@Failure		500				{object}	map[string]string
//	@Router			/dashboard-layouts [get]
func (h *LayoutHandler) List(c *gin.Context) {
	var f dto.LayoutFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	items, total, err := h.uc.List(c.Request.Context(), f)
	if err != nil {
		writeError(c, err)
		return
	}
	writeList(c, items, total)
}

// GetByID godoc
//
//	@Summary		Get a layout placement by id
//	@Tags			Dashboard Layouts
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"Layout placement id"
//	@Success		200	{object}	domain.UtmDashboardVisualization
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/dashboard-layouts/{id} [get]
func (h *LayoutHandler) GetByID(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	res, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	if res == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "layout not found"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// Delete godoc
//
//	@Summary		Delete a layout placement by id
//	@Tags			Dashboard Layouts
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Layout placement id"
//	@Success		200	"Deleted"
//	@Failure		500	{object}	map[string]string
//	@Router			/dashboard-layouts/{id} [delete]
func (h *LayoutHandler) Delete(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	err := h.uc.Delete(c.Request.Context(), id)
	audit.Record(c, audit_connectors.Event{Action: "dashboard.layout.delete", ResourceType: "dashboard_visualization", ResourceID: strconv.FormatUint(id, 10)},
		audit_domain.DASHBOARD_LAYOUT_DELETE_ATTEMPT, audit_domain.DASHBOARD_LAYOUT_DELETE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusOK)
}
