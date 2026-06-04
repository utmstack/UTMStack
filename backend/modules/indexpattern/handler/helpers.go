package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/logger"
)

func writeIPError(c *gin.Context, err error) {
	logger.Error("indexpattern: operation failed: " + err.Error())
	msg := err.Error()
	c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
}

func pathInt64(c *gin.Context, name string) (int64, bool) {
	raw := c.Param(name)
	if raw == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + name})
		return 0, false
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + name + ": " + err.Error()})
		return 0, false
	}
	return v, true
}

func writePagedArray[T any](c *gin.Context, items []T, total int64, page, size int) {
	totalStr := strconv.FormatInt(total, 10)
	c.Header("X-Total-Count", totalStr)

	baseURL := c.Request.URL.Path
	lastPage := int64(0)
	if size > 0 && total > 0 {
		lastPage = (total - 1) / int64(size)
	}
	links := fmt.Sprintf(`<%s?page=0&size=%d>; rel="first"`, baseURL, size)
	links += fmt.Sprintf(`, <%s?page=%d&size=%d>; rel="last"`, baseURL, lastPage, size)
	if int64(page) < lastPage {
		links += fmt.Sprintf(`, <%s?page=%d&size=%d>; rel="next"`, baseURL, page+1, size)
	}
	if page > 0 {
		links += fmt.Sprintf(`, <%s?page=%d&size=%d>; rel="prev"`, baseURL, page-1, size)
	}
	c.Header("Link", links)

	if items == nil {
		items = []T{}
	}
	c.JSON(http.StatusOK, items)
}
