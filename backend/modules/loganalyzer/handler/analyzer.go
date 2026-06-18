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
//	@Param			indexPattern	path		string						true	"Index pattern"
//	@Param			field			path		string						true	"Field to aggregate"
//	@Param			top				path		int							true	"Number of top values"
//	@Param			filters			body		[]common_models.FilterType	false	"Filters to apply"
//	@Success		200				{object}	dto.TopValuesResponse
//	@Failure		400				{object}	map[string]string
//	@Failure		500				{object}	map[string]string
//	@Router			/log-analyzer/top-x-values/{indexPattern}/{field}/{top} [post]
func (h *AnalyzerHandler) TopValues(c *gin.Context) {
	indexPattern := c.Param("indexPattern")
	field := c.Param("field")
	top, _ := strconv.Atoi(c.Param("top"))

	var filters []common_models.FilterType
	// Body is optional; ignore bind errors on an empty body.
	_ = c.ShouldBindJSON(&filters)

	res, err := h.uc.TopValues(c.Request.Context(), indexPattern, field, filters, top)
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
