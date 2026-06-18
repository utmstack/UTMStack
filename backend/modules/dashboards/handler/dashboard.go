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

type DashboardHandler struct{ uc connectors.DashboardUsecase }

func NewDashboardHandler(uc connectors.DashboardUsecase) *DashboardHandler {
	return &DashboardHandler{uc: uc}
}

// Create godoc
//
//	@Summary		Create a dashboard
//	@Description	Creates a new dashboard (board that groups visualizations). The id must be absent.
//	@Tags			Dashboards
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		domain.Dashboard	true	"Dashboard to create"
//	@Success		201		{object}	domain.Dashboard
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/dashboards [post]
func (h *DashboardHandler) Create(c *gin.Context) {
	var d domain.Dashboard
	if err := c.ShouldBindJSON(&d); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.uc.Create(c.Request.Context(), &d, currentUser(c))
	audit.Record(c, audit_connectors.Event{Action: "dashboard.create", ResourceType: "dashboard", ResourceID: d.Name},
		audit_domain.DASHBOARD_CREATE_ATTEMPT, audit_domain.DASHBOARD_CREATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, res)
}

// Update godoc
//
//	@Summary		Update a dashboard
//	@Description	Updates an existing dashboard. The id is required; creation metadata is preserved.
//	@Tags			Dashboards
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		domain.Dashboard	true	"Dashboard to update"
//	@Success		200		{object}	domain.Dashboard
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/dashboards [put]
func (h *DashboardHandler) Update(c *gin.Context) {
	var d domain.Dashboard
	if err := c.ShouldBindJSON(&d); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.uc.Update(c.Request.Context(), &d, currentUser(c))
	audit.Record(c, audit_connectors.Event{Action: "dashboard.update", ResourceType: "dashboard", ResourceID: strconv.FormatUint(d.ID, 10)},
		audit_domain.DASHBOARD_UPDATE_ATTEMPT, audit_domain.DASHBOARD_UPDATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// List godoc
//
//	@Summary		List dashboards
//	@Description	Lists dashboards, optionally filtered by name (substring), paginated.
//	@Tags			Dashboards
//	@Security		BearerAuth
//	@Produce		json
//	@Param			name	query		string	false	"Filter by name (substring)"
//	@Param			page	query		int		false	"Page (0-based)"
//	@Param			size	query		int		false	"Page size"
//	@Success		200		{array}		domain.Dashboard
//	@Header			200		{string}	X-Total-Count	"Total records"
//	@Failure		500		{object}	map[string]string
//	@Router			/dashboards [get]
func (h *DashboardHandler) List(c *gin.Context) {
	var f dto.DashboardFilter
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
//	@Summary		Get a dashboard by id
//	@Tags			Dashboards
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"Dashboard id"
//	@Success		200	{object}	domain.Dashboard
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/dashboards/{id} [get]
func (h *DashboardHandler) GetByID(c *gin.Context) {
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
		c.JSON(http.StatusNotFound, gin.H{"error": "dashboard not found"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// Delete godoc
//
//	@Summary		Delete a dashboard by id
//	@Tags			Dashboards
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Dashboard id"
//	@Success		200	"Deleted"
//	@Failure		500	{object}	map[string]string
//	@Router			/dashboards/{id} [delete]
func (h *DashboardHandler) Delete(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	err := h.uc.Delete(c.Request.Context(), id)
	audit.Record(c, audit_connectors.Event{Action: "dashboard.delete", ResourceType: "dashboard", ResourceID: strconv.FormatUint(id, 10)},
		audit_domain.DASHBOARD_DELETE_ATTEMPT, audit_domain.DASHBOARD_DELETE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusOK)
}
