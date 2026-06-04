package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/notifications/connectors"
	"github.com/utmstack/utmstack/backend/modules/notifications/domain"
	"github.com/utmstack/utmstack/backend/modules/notifications/dto"
)

type NotificationHandler struct {
	usecase connectors.NotificationUsecase
}

func NewNotificationHandler(uc connectors.NotificationUsecase) *NotificationHandler {
	return &NotificationHandler{usecase: uc}
}

// @Summary     Create notification
// @Tags        Notifications
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body dto.CreateNotificationRequest true "Create notification"
// @Success     201 {object} dto.NotificationResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-notifications [post]
func (h *NotificationHandler) Create(c *gin.Context) {
	var req dto.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	n, err := h.usecase.Create(c.Request.Context(), req)
	if err != nil {
		writeNotificationError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.FromEntity(n))
}

// @Summary     List notifications
// @Tags        Notifications
// @Security    BearerAuth
// @Produce     json
// @Param       source query string false "Filter by source"
// @Param       type   query string false "Filter by type"
// @Param       status query string false "Filter by status"
// @Param       read   query bool   false "Filter by read flag"
// @Param       from   query string false "Created at >= (RFC3339)"
// @Param       to     query string false "Created at <= (RFC3339)"
// @Param       page   query int    false "Page (default 1)"
// @Param       size   query int    false "Page size (default 20, max 200)"
// @Param       sort   query string false "field,asc|desc"
// @Success     200 {array} dto.NotificationResponse
// @Header      200 {string} X-Total-Count "Total items"
// @Failure     500 {object} map[string]string
// @Router      /utm-notifications [get]
func (h *NotificationHandler) List(c *gin.Context) {
	q := dto.NotificationListQuery{
		Source: querySource(c, "source"),
		Type:   queryType(c, "type"),
		Status: queryStatus(c, "status"),
		Read:   queryBool(c, "read"),
		From:   queryTime(c, "from"),
		To:     queryTime(c, "to"),
		Page:   queryInt(c, "page", 1),
		Size:   queryInt(c, "size", 20),
		Sort:   c.Query("sort"),
	}
	rows, total, err := h.usecase.List(c.Request.Context(), q)
	if err != nil {
		writeNotificationError(c, err)
		return
	}
	resp := make([]dto.NotificationResponse, len(rows))
	for i := range rows {
		resp[i] = dto.FromEntity(&rows[i])
	}
	writePagedArray(c, resp, total)
}

// @Summary     Get notification by ID
// @Tags        Notifications
// @Security    BearerAuth
// @Produce     json
// @Param       id path int true "Notification ID"
// @Success     200 {object} dto.NotificationResponse
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-notifications/{id} [get]
func (h *NotificationHandler) GetByID(c *gin.Context) {
	id, ok := pathInt64(c, "id")
	if !ok {
		return
	}
	n, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		writeNotificationError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.FromEntity(n))
}

// @Summary     Update read flag
// @Tags        Notifications
// @Security    BearerAuth
// @Produce     json
// @Param       id   path int  true  "Notification ID"
// @Param       read query bool true  "New read flag"
// @Success     200 {object} dto.NotificationResponse
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-notifications/{id}/read [put]
func (h *NotificationHandler) UpdateRead(c *gin.Context) {
	id, ok := pathInt64(c, "id")
	if !ok {
		return
	}
	readStr := c.Query("read")
	if readStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing read query param"})
		return
	}
	read, err := strconv.ParseBool(readStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid read query param"})
		return
	}
	n, err := h.usecase.MarkRead(c.Request.Context(), id, read)
	if err != nil {
		writeNotificationError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.FromEntity(n))
}

// @Summary     Update status
// @Tags        Notifications
// @Security    BearerAuth
// @Produce     json
// @Param       id     path int    true  "Notification ID"
// @Param       status query string true  "New status (ACTIVE|HIDDEN|DELETED)"
// @Success     200 {object} dto.NotificationResponse
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-notifications/{id}/status [put]
func (h *NotificationHandler) UpdateStatus(c *gin.Context) {
	id, ok := pathInt64(c, "id")
	if !ok {
		return
	}
	statusStr := c.Query("status")
	if statusStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing status query param"})
		return
	}
	status := domain.NotificationStatus(statusStr)
	if !status.Valid() {
		writeNotificationError(c, domain.ErrInvalidEnum)
		return
	}
	n, err := h.usecase.UpdateStatus(c.Request.Context(), id, status)
	if err != nil {
		writeNotificationError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.FromEntity(n))
}

// @Summary     Count of unread active notifications
// @Tags        Notifications
// @Security    BearerAuth
// @Produce     json
// @Success     200 {integer} integer
// @Failure     500 {object}  map[string]string
// @Router      /utm-notifications/unread-count [get]
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	n, err := h.usecase.CountUnread(c.Request.Context())
	if err != nil {
		writeNotificationError(c, err)
		return
	}
	c.JSON(http.StatusOK, n)
}

// @Summary     Mark all notifications as read
// @Tags        Notifications
// @Security    BearerAuth
// @Success     204
// @Failure     500 {object} map[string]string
// @Router      /utm-notifications/read-all [put]
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	if err := h.usecase.MarkAllRead(c.Request.Context()); err != nil {
		writeNotificationError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     Delete a notification
// @Tags        Notifications
// @Security    BearerAuth
// @Param       id path int true "Notification ID"
// @Success     204
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /utm-notifications/{id} [delete]
func (h *NotificationHandler) Delete(c *gin.Context) {
	id, ok := pathInt64(c, "id")
	if !ok {
		return
	}
	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		writeNotificationError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
