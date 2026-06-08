package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
)

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrFrameworkNotFound),
		errors.Is(err, domain.ErrControlNotFound),
		errors.Is(err, domain.ErrReportNotFound),
		errors.Is(err, domain.ErrScheduleNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrSystemOwner):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrControlExists), errors.Is(err, domain.ErrFrameworkExists):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrInvalidCron), errors.Is(err, domain.ErrInvalidID):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func pathInt64(c *gin.Context, param string) (int64, bool) {
	v, err := strconv.ParseInt(c.Param(param), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": param + " must be an integer"})
		return 0, false
	}
	return v, true
}

func userIDFromCtx(c *gin.Context) (uint64, bool) {
	raw, ok := c.Get("user_id")
	if !ok {
		return 0, false
	}
	id, ok := raw.(uint64)
	return id, ok
}

func writePagedArray[T any](c *gin.Context, items []T, total int64) {
	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	if items == nil {
		items = []T{}
	}
	c.JSON(http.StatusOK, items)
}
