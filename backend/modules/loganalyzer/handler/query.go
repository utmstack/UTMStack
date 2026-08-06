package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/connectors"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/domain"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/dto"
)

type QueryHandler struct{ uc connectors.QueryUsecase }

func NewQueryHandler(uc connectors.QueryUsecase) *QueryHandler {
	return &QueryHandler{uc: uc}
}

// Create godoc
//
//	@Summary		Create a saved log-analyzer query
//	@Description	Saves a log-explorer search (index pattern + filters + columns). The id must be absent.
//	@Tags			Log Analyzer
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		domain.SavedQuery	true	"Query to create"
//	@Success		201		{object}	domain.SavedQuery
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/log-analyzer/queries [post]
func (h *QueryHandler) Create(c *gin.Context) {
	var q domain.SavedQuery
	if err := c.ShouldBindJSON(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.uc.Create(c.Request.Context(), &q, currentUser(c))
	audit.Record(c, audit_connectors.Event{Action: "log_analyzer.query.create", ResourceType: "log_analyzer_query", ResourceID: q.Name},
		audit_domain.LOG_ANALYZER_QUERY_CREATE_ATTEMPT, audit_domain.LOG_ANALYZER_QUERY_CREATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, res)
}

// Update godoc
//
//	@Summary		Update a saved log-analyzer query
//	@Tags			Log Analyzer
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		domain.SavedQuery	true	"Query to update"
//	@Success		200		{object}	domain.SavedQuery
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/log-analyzer/queries [put]
func (h *QueryHandler) Update(c *gin.Context) {
	var q domain.SavedQuery
	if err := c.ShouldBindJSON(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.uc.Update(c.Request.Context(), &q, currentUser(c))
	audit.Record(c, audit_connectors.Event{Action: "log_analyzer.query.update", ResourceType: "log_analyzer_query", ResourceID: strconv.FormatUint(q.ID, 10)},
		audit_domain.LOG_ANALYZER_QUERY_UPDATE_ATTEMPT, audit_domain.LOG_ANALYZER_QUERY_UPDATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// List godoc
//
//	@Summary		List saved log-analyzer queries
//	@Tags			Log Analyzer
//	@Security		BearerAuth
//	@Produce		json
//	@Param			name	query		string	false	"Filter by name (substring)"
//	@Param			owner	query		string	false	"Filter by owner"
//	@Param			page	query		int		false	"Page (0-based)"
//	@Param			size	query		int		false	"Page size"
//	@Success		200		{array}		domain.SavedQuery
//	@Header			200		{string}	X-Total-Count	"Total records"
//	@Failure		500		{object}	map[string]string
//	@Router			/log-analyzer/queries [get]
func (h *QueryHandler) List(c *gin.Context) {
	var f dto.QueryFilter
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
//	@Summary		Get a saved log-analyzer query by id
//	@Tags			Log Analyzer
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"Query id"
//	@Success		200	{object}	domain.SavedQuery
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/log-analyzer/queries/{id} [get]
func (h *QueryHandler) GetByID(c *gin.Context) {
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
		c.JSON(http.StatusNotFound, gin.H{"error": "query not found"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// Delete godoc
//
//	@Summary		Delete a saved log-analyzer query by id
//	@Tags			Log Analyzer
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Query id"
//	@Success		200	"Deleted"
//	@Failure		500	{object}	map[string]string
//	@Router			/log-analyzer/queries/{id} [delete]
func (h *QueryHandler) Delete(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	err := h.uc.Delete(c.Request.Context(), id)
	audit.Record(c, audit_connectors.Event{Action: "log_analyzer.query.delete", ResourceType: "log_analyzer_query", ResourceID: strconv.FormatUint(id, 10)},
		audit_domain.LOG_ANALYZER_QUERY_DELETE_ATTEMPT, audit_domain.LOG_ANALYZER_QUERY_DELETE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusOK)
}
