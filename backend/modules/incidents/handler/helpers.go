package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	incidenterrors "github.com/utmstack/utmstack/backend/modules/incidents/errors"
	"github.com/utmstack/utmstack/backend/pkg/logger"
)

func writeIncidentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, incidenterrors.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "incident not found"})
	case errors.Is(err, incidenterrors.ErrAlertAlreadyLinked):
		c.JSON(http.StatusConflict, gin.H{"error": "one or more alerts are already linked to an incident"})
	default:
		logger.Error("incident op failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "operation failed"})
	}
}

func writePagedArray[T any](c *gin.Context, items []T, total int64) {
	c.Header("X-Total-Count", strconv.FormatInt(total, 10))

	// Build Link header for pagination (jhipster-style).
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

func queryInt64(c *gin.Context, key string) *int64 {
	v := c.Query(key)
	if v == "" {
		return nil
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return nil
	}
	return &n
}

func queryString(c *gin.Context, key string) *string {
	v := c.Query(key)
	if v == "" {
		return nil
	}
	return &v
}

func queryIntPtr(c *gin.Context, key string) *int {
	v := c.Query(key)
	if v == "" {
		return nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return nil
	}
	return &n
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
