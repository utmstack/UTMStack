package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
)

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrSystemOwned):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrIDForbidden), errors.Is(err, domain.ErrIDRequired), errors.Is(err, domain.ErrNameRequired), errors.Is(err, domain.ErrSpecRequired), errors.Is(err, domain.ErrInvalidSpec):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func pathID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return id, true
}

func writeList[T any](c *gin.Context, items []T, total int64) {
	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	if items == nil {
		items = []T{}
	}
	c.JSON(http.StatusOK, items)
}

func currentUser(c *gin.Context) string { return c.GetString("user_email") }
