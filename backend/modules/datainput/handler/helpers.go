package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	diErrors "github.com/utmstack/utmstack/backend/modules/datainput/errors"
	"github.com/utmstack/utmstack/backend/pkg/logger"
)

func writePagedArray[T any](c *gin.Context, items []T, total int64) {
	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	if items == nil {
		items = []T{}
	}
	c.JSON(http.StatusOK, items)
}

func writeDataInputError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, diErrors.ErrIDExists):
		c.JSON(http.StatusBadRequest, gin.H{"error": "idexists"})
	case errors.Is(err, diErrors.ErrIDNull):
		c.JSON(http.StatusBadRequest, gin.H{"error": "idnull"})
	case errors.Is(err, diErrors.ErrDataInputStatusNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "data input status not found"})
	default:
		logger.Error("datainput: operation failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "operation failed"})
	}
}

func pathStringID(c *gin.Context, name string) (string, bool) {
	id := c.Param(name)
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + name})
		return "", false
	}
	return id, true
}
