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

type VisualizationHandler struct {
	uc connectors.VisualizationUsecase
}

func NewVisualizationHandler(uc connectors.VisualizationUsecase) *VisualizationHandler {
	return &VisualizationHandler{uc: uc}
}

// Create godoc
//
//	@Summary		Create a visualization
//	@Description	Creates a reusable chart definition (query/aggregation + chart type/config). The id must be absent.
//	@Tags			Visualizations
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		domain.UtmVisualization	true	"Visualization to create"
//	@Success		201		{object}	domain.UtmVisualization
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/visualizations [post]
func (h *VisualizationHandler) Create(c *gin.Context) {
	var v domain.UtmVisualization
	if err := c.ShouldBindJSON(&v); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.uc.Create(c.Request.Context(), &v, currentUser(c))
	audit.Record(c, audit_connectors.Event{Action: "visualization.create", ResourceType: "visualization", ResourceID: v.Name},
		audit_domain.VISUALIZATION_CREATE_ATTEMPT, audit_domain.VISUALIZATION_CREATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, res)
}

// Update godoc
//
//	@Summary		Update a visualization
//	@Description	Updates an existing visualization. The id is required; creation metadata is preserved.
//	@Tags			Visualizations
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		domain.UtmVisualization	true	"Visualization to update"
//	@Success		200		{object}	domain.UtmVisualization
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/visualizations [put]
func (h *VisualizationHandler) Update(c *gin.Context) {
	var v domain.UtmVisualization
	if err := c.ShouldBindJSON(&v); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.uc.Update(c.Request.Context(), &v, currentUser(c))
	audit.Record(c, audit_connectors.Event{Action: "visualization.update", ResourceType: "visualization", ResourceID: strconv.FormatUint(v.ID, 10)},
		audit_domain.VISUALIZATION_UPDATE_ATTEMPT, audit_domain.VISUALIZATION_UPDATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// List godoc
//
//	@Summary		List visualizations
//	@Description	Lists visualizations, optionally filtered by name, chart type or index pattern, paginated.
//	@Tags			Visualizations
//	@Security		BearerAuth
//	@Produce		json
//	@Param			name		query		string	false	"Filter by name (substring)"
//	@Param			chartType	query		string	false	"Filter by chart type"
//	@Param			idPattern	query		int		false	"Filter by index pattern id"
//	@Param			page		query		int		false	"Page (0-based)"
//	@Param			size		query		int		false	"Page size"
//	@Success		200			{array}		domain.UtmVisualization
//	@Header			200			{string}	X-Total-Count	"Total records"
//	@Failure		500			{object}	map[string]string
//	@Router			/visualizations [get]
func (h *VisualizationHandler) List(c *gin.Context) {
	var f dto.VisualizationFilter
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
//	@Summary		Get a visualization by id
//	@Tags			Visualizations
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"Visualization id"
//	@Success		200	{object}	domain.UtmVisualization
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/visualizations/{id} [get]
func (h *VisualizationHandler) GetByID(c *gin.Context) {
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
		c.JSON(http.StatusNotFound, gin.H{"error": "visualization not found"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// Delete godoc
//
//	@Summary		Delete a visualization by id
//	@Tags			Visualizations
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Visualization id"
//	@Success		200	"Deleted"
//	@Failure		500	{object}	map[string]string
//	@Router			/visualizations/{id} [delete]
func (h *VisualizationHandler) Delete(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	err := h.uc.Delete(c.Request.Context(), id)
	audit.Record(c, audit_connectors.Event{Action: "visualization.delete", ResourceType: "visualization", ResourceID: strconv.FormatUint(id, 10)},
		audit_domain.VISUALIZATION_DELETE_ATTEMPT, audit_domain.VISUALIZATION_DELETE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusOK)
}
