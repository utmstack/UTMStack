package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	correrrors "github.com/utmstack/utmstack/backend/modules/correlation/errors"
	"github.com/utmstack/utmstack/backend/pkg/logger"
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
	case errors.Is(err, correrrors.ErrIDMustBeAbsent):
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be absent on create"})
	case errors.Is(err, correrrors.ErrIDRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required on update"})
	case errors.Is(err, correrrors.ErrRegexPatternNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": "regex pattern not found"})
	case errors.Is(err, correrrors.ErrRegexPatternSystemOwner):
		c.JSON(http.StatusBadRequest, gin.H{"error": "system regex pattern cannot be modified or deleted"})
	case errors.Is(err, correrrors.ErrTenantConfigNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant config not found"})
	case errors.Is(err, correrrors.ErrDataTypeNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": "data type not found"})
	case errors.Is(err, correrrors.ErrDataTypeSystemOwner):
		c.JSON(http.StatusBadRequest, gin.H{"error": "system data type cannot be modified or deleted"})
	case errors.Is(err, correrrors.ErrCorrelationRuleNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": "correlation rule not found"})
	case errors.Is(err, correrrors.ErrCorrelationRuleSystemOwner):
		c.JSON(http.StatusBadRequest, gin.H{"error": "system's rules can't be updated"})
	case errors.Is(err, correrrors.ErrCorrelationRuleNullDefinition):
		c.JSON(http.StatusBadRequest, gin.H{"error": "correlation rule definition must not be null"})
	case errors.Is(err, correrrors.ErrDataTypesRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": "rule must have at least one data type"})
	default:
		logger.Error("correlation op failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "operation failed"})
	}
}

func isNotFound(err error) bool {
	return errors.Is(err, correrrors.ErrDataTypeNotFound) ||
		errors.Is(err, correrrors.ErrTenantConfigNotFound) ||
		errors.Is(err, correrrors.ErrRegexPatternNotFound) ||
		errors.Is(err, correrrors.ErrCorrelationRuleNotFound)
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
