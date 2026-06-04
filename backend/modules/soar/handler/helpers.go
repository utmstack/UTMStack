package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/pkg/logger"
)

func writePagedArray[T any](c *gin.Context, items []T, total int64) {
	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	if items == nil {
		items = []T{}
	}
	c.JSON(http.StatusOK, items)
}

func writeARRError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrRuleNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "alert response rule not found"})
	case errors.Is(err, domain.ErrTemplateNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "alert response action template not found"})
	case errors.Is(err, domain.ErrRuleNameTaken):
		c.JSON(http.StatusConflict, gin.H{"error": "rule name already in use"})
	case errors.Is(err, domain.ErrIDMustBeAbsent):
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be absent on create"})
	case errors.Is(err, domain.ErrIDRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required on update"})
	default:
		logger.Error("alert response rule op failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "operation failed"})
	}
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

func loginFromCtx(c *gin.Context) string {
	if v, ok := c.Get("user_login"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return "system"
}
