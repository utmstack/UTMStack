package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/logstash/connectors"
	"github.com/utmstack/utmstack/backend/modules/logstash/dto"
)

type FilterHandler struct {
	usecase connectors.FilterUsecase
}

func NewFilterHandler(uc connectors.FilterUsecase) *FilterHandler {
	return &FilterHandler{usecase: uc}
}

// @Summary     Create filter
// @Tags        Logstash Filters
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       pipelineId query int64                  false "Pipeline ID to associate filter with"
// @Param       input      body  dto.CreateFilterRequest true  "Request body"
// @Success     200 {object} dto.FilterResponse
// @Failure     500 {object} map[string]string
// @Router      /utm-filters [post]
func (h *FilterHandler) Create(c *gin.Context) {
	var req dto.CreateFilterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeFilterError(c, err)
		return
	}

	var pipelineID *int64
	if raw := c.Query("pipelineId"); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			writeFilterError(c, err)
			return
		}
		pipelineID = &v
	}

	resp, err := h.usecase.Create(c.Request.Context(), req, pipelineID)
	if err != nil {
		writeFilterError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Update filter
// @Tags        Logstash Filters
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.UpdateFilterRequest true "Request body"
// @Success     200 {object} dto.FilterResponse
// @Failure     500 {object} map[string]string
// @Router      /utm-filters [put]
func (h *FilterHandler) Update(c *gin.Context) {
	var req dto.UpdateFilterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeFilterError(c, err)
		return
	}

	resp, err := h.usecase.Update(c.Request.Context(), req)
	if err != nil {
		writeFilterError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     List filters
// @Tags        Logstash Filters
// @Security    BearerAuth
// @Produce     json
// @Param       id.equals                        query int    false "Filter by ID"
// @Param       filterName.contains              query string false "Filter by name"
// @Param       filterGroupId.equals             query int64  false "Filter by group ID (exact)"
// @Param       filterGroupId.greaterThanOrEqual query int64  false "Filter by group ID (gte)"
// @Param       filterGroupId.lessThanOrEqual    query int64  false "Filter by group ID (lte)"
// @Param       isActive.equals                  query bool   false "Filter by active status"
// @Param       page                             query int    false "Page (0-based)"
// @Param       size                             query int    false "Page size"
// @Success     200 {array} dto.FilterResponse
// @Header      200 {string} X-Total-Count "Total number of items"
// @Failure     500 {object} map[string]string
// @Router      /utm-filters [get]
func (h *FilterHandler) List(c *gin.Context) {
	var f dto.FilterFilters
	if err := c.ShouldBindQuery(&f); err != nil {
		writeFilterError(c, err)
		return
	}

	items, total, err := h.usecase.List(c.Request.Context(), f)
	if err != nil {
		writeFilterError(c, err)
		return
	}
	writePagedArray(c, items, total)
}

// @Summary     Get filters by pipeline ID
// @Tags        Logstash Filters
// @Security    BearerAuth
// @Produce     json
// @Param       pipelineId query int64 true "Pipeline ID"
// @Success     200 {array} dto.FilterResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-filters/by-pipelineid [get]
func (h *FilterHandler) FiltersByPipelineID(c *gin.Context) {
	raw := c.Query("pipelineId")
	if raw == "" {
		logHandlerError("getLogstashFiltersByPipelineId", "pipelineId is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "pipelineId is required"})
		return
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		writeFilterError(c, err)
		return
	}

	items, err := h.usecase.FiltersByPipelineID(c.Request.Context(), id)
	if err != nil {
		writeFilterError(c, err)
		return
	}
	if items == nil {
		items = []dto.FilterResponse{}
	}
	c.JSON(http.StatusOK, items)
}

// @Summary     Get filter by ID
// @Tags        Logstash Filters
// @Security    BearerAuth
// @Produce     json
// @Param       id path int true "Filter ID"
// @Success     200 {object} dto.FilterResponse
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-filters/{id} [get]
func (h *FilterHandler) GetByID(c *gin.Context) {
	id, ok := pathInt64(c, "id")
	if !ok {
		return
	}

	resp, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		writeFilterError(c, err)
		return
	}
	if resp == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "filter not found"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Delete filter
// @Tags        Logstash Filters
// @Security    BearerAuth
// @Param       id path int true "Filter ID"
// @Success     200 "Deleted"
// @Failure     500 {object} map[string]string
// @Router      /utm-filters/{id} [delete]
func (h *FilterHandler) Delete(c *gin.Context) {
	id, ok := pathInt64(c, "id")
	if !ok {
		return
	}

	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		writeFilterError(c, err)
		return
	}
	c.Status(http.StatusOK)
}
