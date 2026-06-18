package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/modules/alerts/dto"
)

type AlertTagHandler struct {
	usecase connectors.AlertTagUsecase
}

func NewAlertTagHandler(uc connectors.AlertTagUsecase) *AlertTagHandler {
	return &AlertTagHandler{usecase: uc}
}

// @Summary     Create alert tag
// @Tags        Alert Tags
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.CreateAlertTagRequest true "Alert tag"
// @Success     201 {object} dto.AlertTagResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-alert-tags [post]
func (h *AlertTagHandler) Create(c *gin.Context) {
	var req dto.CreateAlertTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tag, err := h.usecase.Create(c.Request.Context(), req)
	if err != nil {
		writeAlertError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toAlertTagResponse(tag))
}

// @Summary     Update alert tag
// @Tags        Alert Tags
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.UpdateAlertTagRequest true "Alert tag"
// @Success     200 {object} dto.AlertTagResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-alert-tags [put]
func (h *AlertTagHandler) Update(c *gin.Context) {
	var req dto.UpdateAlertTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tag, err := h.usecase.Update(c.Request.Context(), req)
	if err != nil {
		writeAlertError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAlertTagResponse(tag))
}

// @Summary     List alert tags
// @Tags        Alert Tags
// @Security    BearerAuth
// @Produce     json
// @Param       tagName     query string false "Filter by tag name"
// @Param       systemOwner query bool   false "Filter by system owner"
// @Param       page        query int    false "Page (default 1)"
// @Param       size        query int    false "Page size (default 20)"
// @Success     200 {array} dto.AlertTagResponse
// @Header      200 {string} X-Total-Count "Total items"
// @Failure     500 {object} map[string]string
// @Router      /utm-alert-tags [get]
func (h *AlertTagHandler) List(c *gin.Context) {
	filters := parseAlertTagFilters(c)
	rows, total, err := h.usecase.List(c.Request.Context(), filters)
	if err != nil {
		writeAlertError(c, err)
		return
	}

	responses := make([]dto.AlertTagResponse, 0, len(rows))
	for i := range rows {
		responses = append(responses, toAlertTagResponse(&rows[i]))
	}

	// Return bare array + X-Total-Count (matches Java Spring Page contract).
	writePagedArray(c, responses, total)
}

// @Summary     Get alert tag by ID
// @Tags        Alert Tags
// @Security    BearerAuth
// @Produce     json
// @Param       id path int true "Alert tag ID"
// @Success     200 {object} dto.AlertTagResponse
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-alert-tags/{id} [get]
func (h *AlertTagHandler) GetByID(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	tag, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		writeAlertError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAlertTagResponse(tag))
}

// @Summary     Delete alert tag
// @Tags        Alert Tags
// @Security    BearerAuth
// @Param       id path int true "Alert tag ID"
// @Success     200 "Deleted"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-alert-tags/{id} [delete]
func (h *AlertTagHandler) Delete(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		writeAlertError(c, err)
		return
	}
	// Java returns 200 OK on delete.
	c.Status(http.StatusOK)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func parseAlertTagFilters(c *gin.Context) dto.AlertTagFilters {
	filters := dto.AlertTagFilters{
		Page: queryInt(c, "page", 1),
		Size: queryInt(c, "size", 20),
	}
	if v := c.Query("tagName"); v != "" {
		filters.TagName = &v
	}
	if v := c.Query("systemOwner"); v != "" {
		b := v == "true"
		filters.SystemOwner = &b
	}
	return filters
}

func toAlertTagResponse(tag *domain.UtmAlertTag) dto.AlertTagResponse {
	return dto.AlertTagResponse{
		ID:          tag.ID,
		TagName:     tag.TagName,
		TagColor:    tag.TagColor,
		SystemOwner: tag.SystemOwner,
	}
}
