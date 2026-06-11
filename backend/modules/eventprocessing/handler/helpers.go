package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/domain"
)

func writePagedArray[T any](c *gin.Context, items []T, total int64) {
	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	if items == nil {
		items = []T{}
	}
	c.JSON(http.StatusOK, items)
}

func writeCorrelationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrIDMustBeAbsent):
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be absent on create"})
	case errors.Is(err, domain.ErrIDRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required on update"})
	case errors.Is(err, domain.ErrRegexPatternNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": "regex pattern not found"})
	case errors.Is(err, domain.ErrRegexPatternSystemOwner):
		c.JSON(http.StatusBadRequest, gin.H{"error": "system regex pattern cannot be modified or deleted"})
	case errors.Is(err, domain.ErrTenantConfigNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant config not found"})
	case errors.Is(err, domain.ErrCorrelationRuleNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": "correlation rule not found"})
	case errors.Is(err, domain.ErrCorrelationRuleSystemOwner):
		c.JSON(http.StatusBadRequest, gin.H{"error": "system's rules can't be updated"})
	case errors.Is(err, domain.ErrCorrelationRuleNullDefinition):
		c.JSON(http.StatusBadRequest, gin.H{"error": "correlation rule definition must not be null"})
	case errors.Is(err, domain.ErrCorrelationRuleInvalidContent):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrDataTypesRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": "rule must have at least one data type"})
	default:
		_ = catcher.Error("correlation op failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "operation failed"})
	}
}

func isNotFound(err error) bool {
	return errors.Is(err, domain.ErrTenantConfigNotFound) ||
		errors.Is(err, domain.ErrRegexPatternNotFound) ||
		errors.Is(err, domain.ErrCorrelationRuleNotFound)
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

// --- Logstash filter error mapping (merged from the logstash module) ---

func logHandlerError(operation, msg string) {
	_ = catcher.Error(operation+": "+msg, nil, nil)
}

func writeFilterError(c *gin.Context, err error) {
	if errors.Is(err, domain.ErrFilterNotFound) {
		logHandlerError("logstash filter", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "filter not found"})
		return
	}
	if errors.Is(err, domain.ErrFilterInvalidContent) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, domain.ErrFilterSystemOwner) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "system filter is read-only"})
		return
	}
	logHandlerError("logstash filter", err.Error())
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
