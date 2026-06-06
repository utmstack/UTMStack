package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/logstash/connectors"
	"github.com/utmstack/utmstack/backend/modules/logstash/dto"
)

type FilterGroupHandler struct {
	usecase connectors.FilterGroupUsecase
}

func NewFilterGroupHandler(uc connectors.FilterGroupUsecase) *FilterGroupHandler {
	return &FilterGroupHandler{usecase: uc}
}

// @Summary     Create filter group
// @Tags        Logstash Filter Groups
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.CreateFilterGroupRequest true "Request body"
// @Success     201 {object} dto.FilterGroupResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-logstash-filter-groups [post]
func (h *FilterGroupHandler) Create(c *gin.Context) {
	var req dto.CreateFilterGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.usecase.Create(c.Request.Context(), req)
	if err != nil {
		writeFilterGroupError(c, err)
		return
	}

	c.Header("Location", fmt.Sprintf("/api/v1/utm-logstash-filter-groups/%d", resp.ID))
	c.JSON(http.StatusCreated, resp)
}

// @Summary     Update filter group
// @Tags        Logstash Filter Groups
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.UpdateFilterGroupRequest true "Request body"
// @Success     200 {object} dto.FilterGroupResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-logstash-filter-groups [put]
func (h *FilterGroupHandler) Update(c *gin.Context) {
	var req dto.UpdateFilterGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.usecase.Update(c.Request.Context(), req)
	if err != nil {
		writeFilterGroupError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     List filter groups
// @Tags        Logstash Filter Groups
// @Security    BearerAuth
// @Produce     json
// @Param       id.equals                 query int    false "Filter by ID"
// @Param       groupName.contains        query string false "Filter by group name"
// @Param       groupDescription.contains query string false "Filter by group description"
// @Param       page                      query int    false "Page (0-based)"
// @Param       size                      query int    false "Page size"
// @Success     200 {array} dto.FilterGroupResponse
// @Header      200 {string} X-Total-Count "Total number of items"
// @Failure     500 {object} map[string]string
// @Router      /utm-logstash-filter-groups [get]
func (h *FilterGroupHandler) List(c *gin.Context) {
	var f dto.FilterGroupListFilters
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items, total, err := h.usecase.List(c.Request.Context(), f)
	if err != nil {
		writeFilterGroupError(c, err)
		return
	}
	writePagedArray(c, items, total)
}

// @Summary     Count filter groups
// @Tags        Logstash Filter Groups
// @Security    BearerAuth
// @Produce     json
// @Param       id.equals                 query int    false "Filter by ID"
// @Param       groupName.contains        query string false "Filter by group name"
// @Param       groupDescription.contains query string false "Filter by group description"
// @Success     200 {integer} int64
// @Failure     500 {object} map[string]string
// @Router      /utm-logstash-filter-groups/count [get]
func (h *FilterGroupHandler) Count(c *gin.Context) {
	var f dto.FilterGroupCountFilters
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	count, err := h.usecase.Count(c.Request.Context(), f)
	if err != nil {
		writeFilterGroupError(c, err)
		return
	}
	c.JSON(http.StatusOK, count)
}

// @Summary     Get filter group by ID
// @Tags        Logstash Filter Groups
// @Security    BearerAuth
// @Produce     json
// @Param       id path int true "Filter group ID"
// @Success     200 {object} dto.FilterGroupResponse
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-logstash-filter-groups/{id} [get]
func (h *FilterGroupHandler) GetByID(c *gin.Context) {
	id, ok := pathInt64(c, "id")
	if !ok {
		return
	}

	resp, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		writeFilterGroupError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Delete filter group
// @Tags        Logstash Filter Groups
// @Security    BearerAuth
// @Param       id path int true "Filter group ID"
// @Success     200 "Deleted"
// @Failure     500 {object} map[string]string
// @Router      /utm-logstash-filter-groups/{id} [delete]
func (h *FilterGroupHandler) Delete(c *gin.Context) {
	id, ok := pathInt64(c, "id")
	if !ok {
		return
	}

	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		writeFilterGroupError(c, err)
		return
	}
	c.Status(http.StatusOK)
}
