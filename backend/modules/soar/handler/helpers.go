package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
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
	case errors.Is(err, domain.ErrFlowNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "flow not found"})
	case errors.Is(err, domain.ErrSystemFlowContent):
		c.JSON(http.StatusForbidden, gin.H{"error": "a flow the product ships is read-only"})
	case errors.Is(err, domain.ErrVariableValueRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": "a variable must have a value"})
	case errors.Is(err, domain.ErrRuleNameTaken):
		c.JSON(http.StatusConflict, gin.H{"error": "rule name already in use"})
	case errors.Is(err, domain.ErrSystemRuleReadOnly):
		c.JSON(http.StatusForbidden, gin.H{"error": "system alert response rule is read-only"})
	case errors.Is(err, domain.ErrIDMustBeAbsent):
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be absent on create"})
	case errors.Is(err, domain.ErrIDRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required on update"})
	case errors.Is(err, domain.ErrVariableNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "incident variable not found"})
	case errors.Is(err, domain.ErrIncidentRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "incident record not found"})
	default:
		_ = catcher.Error("alert response rule op failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "operation failed"})
	}
}

func pathUUID(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil || id == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + name})
		return uuid.Nil, false
	}
	return id, true
}

func queryInt(c *gin.Context, name string, def int) int {
	if v := c.Query(name); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func queryString(c *gin.Context, name string) *string {
	if v := c.Query(name); v != "" {
		return &v
	}
	return nil
}

func queryInt64(c *gin.Context, name string) *int64 {
	if v := c.Query(name); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return &n
		}
	}
	return nil
}

func queryIntPtr(c *gin.Context, name string) *int {
	if v := c.Query(name); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return &n
		}
	}
	return nil
}

func queryBoolPtr(c *gin.Context, name string) *bool {
	if v := c.Query(name); v != "" {
		b := v == "true"
		return &b
	}
	return nil
}

func loginFromCtx(c *gin.Context) string {
	if v, ok := c.Get("user_email"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return "system"
}
