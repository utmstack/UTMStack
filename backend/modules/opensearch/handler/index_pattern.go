package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/opensearch/connectors"
	"github.com/utmstack/utmstack/backend/modules/opensearch/dto"
)

type IndexPatternHandler struct {
	usecase connectors.IndexPatternUsecase
}

func NewIndexPatternHandler(uc connectors.IndexPatternUsecase) *IndexPatternHandler {
	return &IndexPatternHandler{usecase: uc}
}

// @Summary     Create index pattern
// @Tags        Index Patterns
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.CreateIndexPatternRequest true "Index pattern to create (id must be absent)"
// @Success     200 {object} dto.IndexPatternResponse
// @Failure     500 {object} map[string]string
// @Router      /utm-index-patterns [post]
func (h *IndexPatternHandler) Create(c *gin.Context) {
	var req dto.CreateIndexPatternRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.usecase.Create(c.Request.Context(), req)
	audit.Record(c, audit_connectors.Event{Action: "index.pattern.create", ResourceType: "index_pattern", ResourceID: req.Pattern},
		audit_domain.INDEX_PATTERN_CREATE_ATTEMPT, audit_domain.INDEX_PATTERN_CREATE_SUCCESS, err)
	if err != nil {
		writeIPError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Update index pattern
// @Tags        Index Patterns
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.UpdateIndexPatternRequest true "Index pattern to update (id required)"
// @Success     200 {object} dto.IndexPatternResponse
// @Failure     500 {object} map[string]string
// @Router      /utm-index-patterns [put]
func (h *IndexPatternHandler) Update(c *gin.Context) {
	var req dto.UpdateIndexPatternRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.usecase.Update(c.Request.Context(), req)
	audit.Record(c, audit_connectors.Event{Action: "index.pattern.update", ResourceType: "index_pattern", ResourceID: strconv.FormatInt(req.ID, 10)},
		audit_domain.INDEX_PATTERN_UPDATE_ATTEMPT, audit_domain.INDEX_PATTERN_UPDATE_SUCCESS, err)
	if err != nil {
		writeIPError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     List index patterns
// @Tags        Index Patterns
// @Security    BearerAuth
// @Produce     json
// @Param       id.equals             query int    false "Filter by exact id"
// @Param       id.greaterThan        query int    false "Filter by id greater than"
// @Param       id.lessThan           query int    false "Filter by id less than"
// @Param       pattern.contains      query string false "Filter by pattern containing value"
// @Param       pattern.equals        query string false "Filter by exact pattern"
// @Param       patternModule.contains query string false "Filter by patternModule containing value"
// @Param       patternSystem.equals  query bool   false "Filter by patternSystem"
// @Param       isActive.equals       query bool   false "Filter by isActive"
// @Param       page                  query int    false "Page (0-based)"
// @Param       size                  query int    false "Page size"
// @Success     200 {array} dto.IndexPatternResponse
// @Header      200 {string} X-Total-Count "Total items"
// @Header      200 {string} Link "RFC-5988 pagination links"
// @Failure     500 {object} map[string]string
// @Router      /utm-index-patterns [get]
func (h *IndexPatternHandler) List(c *gin.Context) {
	var f dto.IndexPatternFilters
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items, total, err := h.usecase.List(c.Request.Context(), f)
	if err != nil {
		writeIPError(c, err)
		return
	}
	writePagedArray(c, items, total, f.Page, f.Size)
}

// @Summary     List index patterns with fields
// @Tags        Index Patterns
// @Security    BearerAuth
// @Produce     json
// @Param       id.equals             query int    false "Filter by exact id"
// @Param       id.greaterThan        query int    false "Filter by id greater than"
// @Param       id.lessThan           query int    false "Filter by id less than"
// @Param       pattern.contains      query string false "Filter by pattern containing value"
// @Param       pattern.equals        query string false "Filter by exact pattern"
// @Param       patternModule.contains query string false "Filter by patternModule containing value"
// @Param       patternSystem.equals  query bool   false "Filter by patternSystem"
// @Param       isActive.equals       query bool   false "Filter by isActive"
// @Param       page                  query int    false "Page (0-based)"
// @Param       size                  query int    false "Page size"
// @Success     200 {array} dto.IndexPatternFieldsResponse
// @Header      200 {string} X-Total-Count "Total items"
// @Header      200 {string} Link "RFC-5988 pagination links"
// @Failure     500 {object} map[string]string
// @Router      /utm-index-patterns/fields [get]
func (h *IndexPatternHandler) Fields(c *gin.Context) {
	var f dto.IndexPatternFilters
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items, total, err := h.usecase.Fields(c.Request.Context(), f)
	if err != nil {
		writeIPError(c, err)
		return
	}
	writePagedArray(c, items, total, f.Page, f.Size)
}

// @Summary     Get index pattern by ID
// @Tags        Index Patterns
// @Security    BearerAuth
// @Produce     json
// @Param       id path int true "Index pattern ID"
// @Success     200 {object} dto.IndexPatternResponse
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-index-patterns/{id} [get]
func (h *IndexPatternHandler) GetByID(c *gin.Context) {
	id, ok := pathInt64(c, "id")
	if !ok {
		return
	}

	resp, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		writeIPError(c, err)
		return
	}
	if resp == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "index pattern not found"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Delete index pattern
// @Tags        Index Patterns
// @Security    BearerAuth
// @Param       id path int true "Index pattern ID"
// @Success     200 "Deleted"
// @Failure     500 {object} map[string]string
// @Router      /utm-index-patterns/{id} [delete]
func (h *IndexPatternHandler) Delete(c *gin.Context) {
	id, ok := pathInt64(c, "id")
	if !ok {
		return
	}

	err := h.usecase.Delete(c.Request.Context(), id)
	audit.Record(c, audit_connectors.Event{Action: "index.pattern.delete", ResourceType: "index_pattern", ResourceID: strconv.FormatInt(id, 10)},
		audit_domain.INDEX_PATTERN_DELETE_ATTEMPT, audit_domain.INDEX_PATTERN_DELETE_SUCCESS, err)
	if err != nil {
		writeIPError(c, err)
		return
	}
	c.Status(http.StatusOK)
}
