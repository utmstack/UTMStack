package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	lsErrors "github.com/utmstack/utmstack/backend/modules/logstash/errors"
	"github.com/utmstack/utmstack/backend/pkg/logger"
)

// logHandlerError logs an operation error the same way Java's ApplicationEventService does.
// When ApplicationEventService is ported, replace this with the real service call.
func logHandlerError(operation, msg string) {
	logger.Error(operation + ": " + msg)
}

func writeFilterGroupError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, lsErrors.ErrFilterGroupIDExists):
		logHandlerError("logstash filter group", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "idexists"})
	case errors.Is(err, lsErrors.ErrFilterGroupIDNull):
		logHandlerError("logstash filter group", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "idnull"})
	case errors.Is(err, lsErrors.ErrFilterGroupNotFound):
		logHandlerError("logstash filter group", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "filter group not found"})
	default:
		logHandlerError("logstash filter group", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "operation failed"})
	}
}

func writeFilterError(c *gin.Context, err error) {
	if errors.Is(err, lsErrors.ErrFilterNotFound) {
		logHandlerError("logstash filter", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "filter not found"})
		return
	}
	logHandlerError("logstash filter", err.Error())
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func pathInt64(c *gin.Context, name string) (int64, bool) {
	raw := c.Param(name)
	if raw == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + name})
		return 0, false
	}
	val, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + name})
		return 0, false
	}
	return val, true
}

func writePagedArray[T any](c *gin.Context, items []T, total int64) {
	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	if items == nil {
		items = []T{}
	}
	c.JSON(http.StatusOK, items)
}
