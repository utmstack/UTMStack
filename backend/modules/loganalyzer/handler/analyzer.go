package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/connectors"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type AnalyzerHandler struct{ uc connectors.AnalyzerUsecase }

func NewAnalyzerHandler(uc connectors.AnalyzerUsecase) *AnalyzerHandler {
	return &AnalyzerHandler{uc: uc}
}

// TopValues godoc
//
//	@Summary		Top-N values of a field
//	@Description	Returns the most frequent values of a field over an index (terms aggregation) plus the total doc count. Body is an optional filter list.
//	@Tags			Log Analyzer
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			dataset	path		string						true	"Index pattern"
//	@Param			field			path		string						true	"Field to aggregate"
//	@Param			top				path		int							true	"Number of top values"
//	@Param			filters			body		[]common_models.FilterType	false	"Filters to apply"
//	@Success		200				{object}	dto.TopValuesResponse
//	@Failure		400				{object}	map[string]string
//	@Failure		500				{object}	map[string]string
//	@Router			/log-analyzer/top-x-values/{dataset}/{field}/{top} [post]
func (h *AnalyzerHandler) TopValues(c *gin.Context) {
	dataset := c.Param("dataset")
	field := c.Param("field")
	top, _ := strconv.Atoi(c.Param("top"))

	var filters []common_models.FilterType
	// Body is optional; ignore bind errors on an empty body.
	_ = c.ShouldBindJSON(&filters)

	// The data type is optional and comes on the query string: a caller asking
	// about one kind of record inside the dataset, which an index pattern used
	// to say by its name.
	res, err := h.uc.TopValues(c.Request.Context(), dataset, c.Query("dataType"), field, filters, top)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// ChartView godoc
//
//	@Summary		Chart/timeline aggregation
//	@Description	Returns category/value arrays for the explorer chart — a date_histogram when `interval` is set, otherwise a terms aggregation over `field`.
//	@Tags			Log Analyzer
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		dto.ChartViewRequest	true	"Chart view request"
//	@Success		200		{object}	dto.ChartViewResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/log-analyzer/chart-view [post]
func (h *AnalyzerHandler) ChartView(c *gin.Context) {
	var req dto.ChartViewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.uc.ChartView(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// Datasets godoc
//
//	@Summary		What can be explored
//	@Tags			Log Analyzer
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{array}	string
//	@Router			/log-analyzer/datasets [get]
func (h *AnalyzerHandler) Datasets(c *gin.Context) {
	c.JSON(http.StatusOK, h.uc.Datasets())
}

// Fields godoc
//
//	@Summary		What can be asked of a dataset
//	@Description	The columns it has, from the store itself rather than a registry — so a field that exists is offered and one that does not is not.
//	@Tags			Log Analyzer
//	@Security		BearerAuth
//	@Produce		json
//	@Param			dataset	path		string	true	"logs | alerts"
//	@Success		200		{array}		dto.Field
//	@Failure		400		{object}	map[string]string
//	@Router			/log-analyzer/datasets/{dataset}/fields [get]
func (h *AnalyzerHandler) Fields(c *gin.Context) {
	fields, err := h.uc.Fields(c.Request.Context(), c.Param("dataset"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, fields)
}

// Search godoc
//
//	@Summary		Search documents
//	@Description	A page of documents for the explorer, read through the event store rather than the index gateway.
//	@Tags			Log Analyzer
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		dto.SearchRequest	true	"Search request"
//	@Success		200		{object}	dto.SearchResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/log-analyzer/search [post]
func (h *AnalyzerHandler) Search(c *gin.Context) {
	var req dto.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.uc.Search(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// DataTypes godoc
//
//	@Summary		Data types present in a dataset
//	@Tags			Log Analyzer
//	@Security		BearerAuth
//	@Produce		json
//	@Param			dataset	path	string	true	"Dataset"
//	@Success		200	{array}	string
//	@Router			/log-analyzer/datasets/{dataset}/data-types [get]
func (h *AnalyzerHandler) DataTypes(c *gin.Context) {
	out, err := h.uc.DataTypes(c.Request.Context(), c.Param("dataset"))
	if err != nil {
		writeError(c, err)
		return
	}
	if out == nil {
		out = []string{}
	}
	c.JSON(http.StatusOK, out)
}

type sqlSearchRequest struct {
	Query string `json:"query" binding:"required"`
}

// SearchSQL runs a caller-written SELECT over the scoped datasets.
//
//	@Summary		Search with SQL
//	@Description	Runs a SELECT over the `logs` and `alerts` datasets. Both names resolve to the caller's own tenant; naming a real table, another database or a table function is refused.
//	@Tags			Log Analyzer
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			page	query		int					false	"Page (default 1)"
//	@Param			size	query		int					false	"Page size (default 50, max 500)"
//	@Param			input	body		sqlSearchRequest	true	"SQL"
//	@Success		200		{object}	dto.SearchResponse
//	@Failure		400		{object}	map[string]string
//	@Router			/log-analyzer/search-sql [post]
func (h *AnalyzerHandler) SearchSQL(c *gin.Context) {
	var req sqlSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	page, _ := strconv.Atoi(c.Query("page"))
	size, _ := strconv.Atoi(c.Query("size"))

	res, err := h.uc.SearchSQL(c.Request.Context(), req.Query, page, size)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}
