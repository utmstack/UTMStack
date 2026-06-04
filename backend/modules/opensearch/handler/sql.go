package handler

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/opensearch/dto"
	"github.com/utmstack/utmstack/backend/modules/opensearch/usecase"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
	"github.com/utmstack/utmstack/backend/pkg/logger"
)

type sqlUsecase interface {
	SearchSQL(ctx context.Context, sqlDto dto.SqlSearchDto, page, size int) ([]map[string]any, int64, error)
	SearchCSV(ctx context.Context, filters []common_models.FilterType, top int, indexPattern string, columns []dto.DataColumn) ([]map[string]any, error)
	ExportCSV(hits []map[string]any, columns []dto.DataColumn, w io.Writer) error
}

type SQLHandler struct {
	uc sqlUsecase
}

func NewSQLHandler(uc sqlUsecase) *SQLHandler {
	return &SQLHandler{uc: uc}
}

// @Summary     Search using SQL
// @Description Executes a validated SQL query against the OpenSearch SQL plugin with automatic pagination.
// @Tags        OpenSearch
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       page query int false "Page number (1-based)"
// @Param       size query int false "Page size"
// @Param       body body dto.SqlSearchDto true "SQL query"
// @Success     200 {array}  map[string]interface{}
// @Header      200 {string} X-Total-Count "Total rows (capped at 10000)"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /elasticsearch/search/sql [post]
func (h *SQLHandler) SearchSQL(c *gin.Context) {
	page := queryIntDefault(c, "page", 1)
	size := queryIntDefault(c, "size", 10)

	var body dto.SqlSearchDto
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if strings.TrimSpace(body.Query) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
		return
	}

	if err := usecase.ValidateSQL(body.Query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hits, total, err := h.uc.SearchSQL(c.Request.Context(), body, page, size)
	if err != nil {
		writeOSError(c, err)
		return
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, hits)
}

// @Summary     Export search results to CSV
// @Description Fetches documents matching the request body filters and streams them as a CSV download.
// @Tags        OpenSearch
// @Security    BearerAuth
// @Accept      json
// @Produce     text/csv
// @Param       body body dto.CsvExportingParams true "CSV export parameters"
// @Success     200 {string} string "CSV file stream"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /elasticsearch/search/csv [post]
func (h *SQLHandler) SearchCSV(c *gin.Context) {
	var params dto.CsvExportingParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hits, err := h.uc.SearchCSV(c.Request.Context(), params.Filters, params.Top, params.IndexPattern, params.Columns)
	if err != nil {
		writeOSError(c, err)
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", `attachment; filename="export.csv"`)
	c.Status(http.StatusOK)

	if err := h.uc.ExportCSV(hits, params.Columns, c.Writer); err != nil {
		logger.Error("SearchCSV: ExportCSV failed: " + err.Error())
		// Headers already sent; write a plain-text error trailer so callers can detect it.
		c.Writer.WriteHeader(http.StatusInternalServerError)
	}
}
