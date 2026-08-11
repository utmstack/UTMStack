package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/usecase"
)

type IngestionStatsHandler struct {
	uc connectors.IngestionStatsUsecase
}

func NewIngestionStatsHandler(uc connectors.IngestionStatsUsecase) *IngestionStatsHandler {
	return &IngestionStatsHandler{uc: uc}
}

// Totals godoc
//
//	@Summary     Ingestion totals by dataSource or dataType
//	@Description Sums the statistics the stats plugin writes: how many events
//	             arrived and how many bytes they weighed, grouped by dataSource
//	             or dataType over a window (default the last 24 hours).
//	@Tags        Event Processing Stats
//	@Security    BearerAuth
//	@Produce     json
//	@Param       groupBy query string true  "dataSource | dataType"
//	@Param       status  query string false "received | parsing_dropped | analysis_dropped | correlation_dropped | all (default received)"
//	@Param       from    query string false "RFC3339 start (default now-24h)"
//	@Param       to      query string false "RFC3339 end (default now)"
//	@Param       top     query int    false "Max buckets (default 100)"
//	@Success     200 {object} dto.IngestionStatsResponse
//	@Failure     400 {object} map[string]string
//	@Failure     500 {object} map[string]string
//	@Router      /eventprocessing/ingestion-stats [get]
func (h *IngestionStatsHandler) Totals(c *gin.Context) {
	resp, err := h.uc.Totals(
		c.Request.Context(),
		c.Query("groupBy"),
		c.Query("status"),
		c.Query("from"),
		c.Query("to"),
		queryIntDefault(c, "top", 0),
	)
	if err != nil {
		writeStatsError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Timeline godoc
//
//	@Summary     Ingestion time series
//	@Description Date-histogram of received events over a window (default 24h).
//	             Without groupBy returns a single line; with groupBy returns one
//	             line per dataSource/dataType.
//	@Tags        Event Processing Stats
//	@Security    BearerAuth
//	@Produce     json
//	@Param       groupBy  query string false "dataSource | dataType (omit for a single line)"
//	@Param       status   query string false "received (default) | *_dropped | all"
//	@Param       interval query string false "auto (default) | 5m | 1h | 1d | ..."
//	@Param       from     query string false "RFC3339 start (default now-24h)"
//	@Param       to       query string false "RFC3339 end (default now)"
//	@Param       top      query int    false "Max series when groupBy set (default 100)"
//	@Param       dataSource query string false "Scope the timeline to a single source (by name)"
//	@Success     200 {object} dto.IngestionTimelineResponse
//	@Failure     400 {object} map[string]string
//	@Failure     500 {object} map[string]string
//	@Router      /eventprocessing/ingestion-stats/timeline [get]
func (h *IngestionStatsHandler) Timeline(c *gin.Context) {
	resp, err := h.uc.Timeline(
		c.Request.Context(),
		c.Query("groupBy"),
		c.Query("status"),
		c.Query("interval"),
		c.Query("from"),
		c.Query("to"),
		queryIntDefault(c, "top", 0),
		c.Query("dataSource"),
	)
	if err != nil {
		writeStatsError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func writeStatsError(c *gin.Context, err error) {
	if errors.Is(err, usecase.ErrInvalidStatsParam) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_ = catcher.Error("ingestion stats query failed", err, nil)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "operation failed"})
}

func queryIntDefault(c *gin.Context, name string, def int) int {
	if v := c.Query(name); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
