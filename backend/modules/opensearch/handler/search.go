package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/opensearch/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type searchUsecase interface {
	Search(ctx context.Context, filters []common_models.FilterType, top int, indexPattern string, includeChildren bool, page, size int, sortBy, sortOrder string) ([]map[string]any, int64, error)
	GenericSearch(ctx context.Context, index string, filters []common_models.FilterType, top int, page, size int) ([]map[string]any, int64, error)
	Count(ctx context.Context, filters []common_models.FilterType, indexPattern string) (bool, error)
}

type SearchHandler struct {
	uc searchUsecase
}

func NewSearchHandler(uc searchUsecase) *SearchHandler {
	return &SearchHandler{uc: uc}
}

// @Summary     Search documents
// @Description Executes a filter-based search against an index pattern with optional pagination and children enrichment.
// @Tags        OpenSearch
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       top             query int    false "Maximum number of results (default 500)"
// @Param       indexPattern    query string true  "Index pattern"
// @Param       includeChildren query bool   false "Enrich each hit with hasChildren, echoes, and last_echo fields"
// @Param       page            query int    false "Page number (1-based)"
// @Param       size            query int    false "Page size"
// @Param       filters         body  []common_models.FilterType false "List of filters"
// @Success     200 {array}  map[string]interface{}
// @Header      200 {string} X-Total-Count "Total matching documents (capped at top)"
// @Failure     500 {object} map[string]string
// @Router      /elasticsearch/search [post]
func (h *SearchHandler) Search(c *gin.Context) {
	top := queryIntDefault(c, "top", 500)
	indexPattern := c.Query("indexPattern")
	includeChildren := queryBoolDefault(c, "includeChildren", false)
	page := queryIntDefault(c, "page", 1)
	size := queryIntDefault(c, "size", 10)
	sortBy := c.Query("sortBy")
	sortOrder := c.DefaultQuery("sortOrder", "desc")

	var filters []common_models.FilterType
	_ = c.ShouldBindJSON(&filters)

	hits, total, err := h.uc.Search(c.Request.Context(), filters, top, indexPattern, includeChildren, page, size, sortBy, sortOrder)
	if err != nil {
		writeOSError(c, err)
		return
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, hits)
}

// @Summary     Generic search
// @Description Executes a filter-based search against a single named index (no wildcard pattern).
// @Tags        OpenSearch
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       page query int false "Page number (1-based)"
// @Param       size query int false "Page size"
// @Param       body body dto.GenericSearchRequest true "Search request"
// @Success     200 {array}  map[string]interface{}
// @Header      200 {string} X-Total-Count "Total matching documents (capped at top)"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /elasticsearch/generic-search [post]
func (h *SearchHandler) GenericSearch(c *gin.Context) {
	var req dto.GenericSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	page := queryIntDefault(c, "page", 1)
	size := queryIntDefault(c, "size", 10)

	hits, total, err := h.uc.GenericSearch(c.Request.Context(), req.Index, req.Filters, req.Top, page, size)
	if err != nil {
		writeOSError(c, err)
		return
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, hits)
}

// @Summary     Check document existence
// @Description Returns true if at least one document in the index pattern matches the given filters.
// @Tags        OpenSearch
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       indexPattern query string false "Index pattern"
// @Param       filters      body  []common_models.FilterType false "List of filters"
// @Success     200 {boolean} bool
// @Failure     500 {object} map[string]string
// @Router      /elasticsearch/count [post]
func (h *SearchHandler) Count(c *gin.Context) {
	indexPattern := c.Query("indexPattern")

	var filters []common_models.FilterType
	_ = c.ShouldBindJSON(&filters)

	exists, err := h.uc.Count(c.Request.Context(), filters, indexPattern)
	if err != nil {
		writeOSError(c, err)
		return
	}

	c.JSON(http.StatusOK, exists)
}

// ---------------------------------------------------------------------------
// local query param helpers
// ---------------------------------------------------------------------------

func queryIntDefault(c *gin.Context, key string, def int) int {
	v := c.Query(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return def
	}
	return n
}

func queryBoolDefault(c *gin.Context, key string, def bool) bool {
	v := c.Query(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}
