package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/logger"
)

func writeCollectorError(c *gin.Context, err error) {
	logger.Error("collectors: " + err.Error())
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func writePagedResponse[T any](c *gin.Context, items []T, total int32) {
	c.Header("X-Total-Count", fmt.Sprintf("%d", total))
	if items == nil {
		items = []T{}
	}
	c.JSON(http.StatusOK, items)
}

func pathInt64(c *gin.Context, name string) (int64, bool) {
	raw := c.Param(name)
	if raw == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing path parameter: " + name})
		return 0, false
	}
	val, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path parameter: " + name})
		return 0, false
	}
	return val, true
}
