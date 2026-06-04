package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/notifications/domain"
	notiferrors "github.com/utmstack/utmstack/backend/modules/notifications/errors"
	"github.com/utmstack/utmstack/backend/pkg/logger"
)

func writeNotificationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, notiferrors.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
	case errors.Is(err, notiferrors.ErrInvalidEnum):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid enum value"})
	default:
		logger.Error("notification op failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "operation failed"})
	}
}

func writePagedArray[T any](c *gin.Context, items []T, total int64) {
	c.Header("X-Total-Count", strconv.FormatInt(total, 10))

	page := queryInt(c, "page", 1)
	size := queryInt(c, "size", 20)
	if size < 1 {
		size = 20
	}
	totalPages := int((total + int64(size) - 1) / int64(size))
	path := c.Request.URL.Path
	var links []string
	if page < totalPages {
		links = append(links, fmt.Sprintf("<%s?page=%d&size=%d>; rel=\"next\"", path, page+1, size))
	}
	if page > 1 {
		links = append(links, fmt.Sprintf("<%s?page=%d&size=%d>; rel=\"prev\"", path, page-1, size))
	}
	if len(links) > 0 {
		linkVal := ""
		for i, l := range links {
			if i > 0 {
				linkVal += ", "
			}
			linkVal += l
		}
		c.Header("Link", linkVal)
	}

	if items == nil {
		items = []T{}
	}
	c.JSON(http.StatusOK, items)
}

func pathInt64(c *gin.Context, name string) (int64, bool) {
	idStr := c.Param(name)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + name})
		return 0, false
	}
	return id, true
}

func queryInt(c *gin.Context, key string, defaultVal int) int {
	v := c.Query(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return defaultVal
	}
	return n
}

func queryBool(c *gin.Context, key string) *bool {
	v := c.Query(key)
	if v == "" {
		return nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return nil
	}
	return &b
}

func queryTime(c *gin.Context, key string) *time.Time {
	v := c.Query(key)
	if v == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return nil
	}
	return &t
}

func querySource(c *gin.Context, key string) *domain.NotificationSource {
	v := c.Query(key)
	if v == "" {
		return nil
	}
	s := domain.NotificationSource(v)
	if !s.Valid() {
		return nil
	}
	return &s
}

func queryType(c *gin.Context, key string) *domain.NotificationType {
	v := c.Query(key)
	if v == "" {
		return nil
	}
	t := domain.NotificationType(v)
	if !t.Valid() {
		return nil
	}
	return &t
}

func queryStatus(c *gin.Context, key string) *domain.NotificationStatus {
	v := c.Query(key)
	if v == "" {
		return nil
	}
	s := domain.NotificationStatus(v)
	if !s.Valid() {
		return nil
	}
	return &s
}
