package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
)

// writePagedArray writes the items as a bare JSON array and sets
// the X-Total-Count header, matching the Java Spring Page response contract.
func writePagedArray[T any](c *gin.Context, items []T, total int64) {
	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	if items == nil {
		items = []T{}
	}
	c.JSON(http.StatusOK, items)
}

// isNotFoundErr returns true for any "not found" sentinel error in the alerts domain.
func isNotFoundErr(err error) bool {
	return errors.Is(err, domain.ErrAlertTagNotFound) ||
		errors.Is(err, domain.ErrAlertTagRuleNotFound)
}

// queryInt reads a positive integer query param, falling back to defaultVal
// when absent, non-numeric, or zero.
func queryInt(c *gin.Context, key string, defaultVal int) int {
	v := c.Query(key)
	if v == "" {
		return defaultVal
	}
	n := 0
	for _, ch := range v {
		if ch < '0' || ch > '9' {
			return defaultVal
		}
		n = n*10 + int(ch-'0')
	}
	if n == 0 {
		return defaultVal
	}
	return n
}

func pathID(c *gin.Context, name string) (uint64, bool) {
	idStr := c.Param(name)
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + name})
		return 0, false
	}
	return id, true
}

func writeAlertError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrAlertTagNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "alert tag not found"})
	case errors.Is(err, domain.ErrAlertTagRuleNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "alert tag rule not found"})
	case errors.Is(err, domain.ErrTagNameTaken):
		c.JSON(http.StatusConflict, gin.H{"error": "tag name already in use"})
	case errors.Is(err, domain.ErrRuleNameTaken):
		c.JSON(http.StatusConflict, gin.H{"error": "rule name already in use"})
	case errors.Is(err, domain.ErrInvalidAlertStatus):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert status"})
	case errors.Is(err, domain.ErrMissingAlertID):
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing alert id"})
	default:
		_ = catcher.Error("alert op failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "operation failed"})
	}
}
