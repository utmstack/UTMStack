package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/datasources/domain"
)

func pathID(c *gin.Context, name string) (uint64, bool) {
	idStr := c.Param(name)
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + name})
		return 0, false
	}
	return id, true
}

// writeError maps usecase/repository errors to HTTP responses uniformly.
func writeError(c *gin.Context, op string, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	case errors.Is(err, domain.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
	case errors.Is(err, domain.ErrUnknownProperty):
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown property"})
	case errors.Is(err, domain.ErrProbeNotConfigured):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "probe url not configured"})
	case errors.Is(err, domain.ErrProbeUnavailable):
		c.JSON(http.StatusBadGateway, gin.H{"error": "probe unavailable"})
	default:
		_ = catcher.Error(op+" failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": op + " failed"})
	}
}
