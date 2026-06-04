package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	ospkg "github.com/utmstack/utmstack/backend/pkg/opensearch"
)

type indexUsecase interface {
	IndexProperties(ctx context.Context, pattern string) ([]ospkg.IndexPropertyType, error)
	IndexAll(ctx context.Context, pattern string, includeSystem bool, page, size int) ([]ospkg.IndexInfo, int64, error)
	DeleteIndex(ctx context.Context, indices []string) error
}

type IndexHandler struct {
	uc indexUsecase
}

func NewIndexHandler(uc indexUsecase) *IndexHandler {
	return &IndexHandler{uc: uc}
}

// @Summary     Get index properties
// @Description Returns a flattened list of field name + type for the given index pattern.
// @Tags        OpenSearch
// @Security    BearerAuth
// @Produce     json
// @Param       indexPattern query string true "Index pattern"
// @Success     200 {array}  ospkg.IndexPropertyType
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /elasticsearch/index/properties [get]
func (h *IndexHandler) Properties(c *gin.Context) {
	pattern := c.Query("indexPattern")
	if pattern == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "indexPattern is required"})
		return
	}

	props, err := h.uc.IndexProperties(c.Request.Context(), pattern)
	if err != nil {
		writeOSError(c, err)
		return
	}
	if props == nil {
		props = []ospkg.IndexPropertyType{}
	}
	c.JSON(http.StatusOK, props)
}

// @Summary     List all indices
// @Description Returns a paginated list of OpenSearch indices, optionally filtered by pattern.
// @Tags        OpenSearch
// @Security    BearerAuth
// @Produce     json
// @Param       includeSystemIndex query bool   false "Include system indices (starting with '.')"
// @Param       pattern            query string false "Filter by index name pattern"
// @Param       page               query int    false "Page number (1-based)"
// @Param       size               query int    false "Page size"
// @Success     200 {array}  ospkg.IndexInfo
// @Header      200 {string} X-Total-Count "Total number of matching indices"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /elasticsearch/index/all [get]
func (h *IndexHandler) IndexAll(c *gin.Context) {
	includeSystem := false
	if v := c.Query("includeSystemIndex"); v != "" {
		parsed, err := strconv.ParseBool(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid includeSystemIndex: %s", v)})
			return
		}
		includeSystem = parsed
	}

	pattern := c.Query("pattern")

	page := 1
	if v := c.Query("page"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil || parsed < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "page must be a positive integer"})
			return
		}
		page = parsed
	}

	size := 10
	if v := c.Query("size"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil || parsed < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "size must be a positive integer"})
			return
		}
		size = parsed
	}

	indices, total, err := h.uc.IndexAll(c.Request.Context(), pattern, includeSystem, page, size)
	if err != nil {
		writeOSError(c, err)
		return
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, indices)
}

// @Summary     Delete indices
// @Description Deletes one or more OpenSearch indices by name.
// @Tags        OpenSearch
// @Security    BearerAuth
// @Accept      json
// @Param       indices body []string true "List of index names to delete"
// @Success     200 "Deleted"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /elasticsearch/index/delete-index [post]
func (h *IndexHandler) DeleteIndex(c *gin.Context) {
	var indices []string
	if err := c.ShouldBindJSON(&indices); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.uc.DeleteIndex(c.Request.Context(), indices); err != nil {
		writeOSError(c, err)
		return
	}

	c.Status(http.StatusOK)
}
