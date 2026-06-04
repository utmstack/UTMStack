package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
)

// writeError maps domain sentinels to HTTP statuses. Anything that's not a
// recognised sentinel surfaces as 500 with a logged trace so we don't leak
// internal details.
func writeError(c *gin.Context, op string, err error) {
	switch {
	case errors.Is(err, domain.ErrModuleNotFound),
		errors.Is(err, domain.ErrGroupNotFound),
		errors.Is(err, domain.ErrConfigNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrModuleKindNotPorted):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrInvalidActivation),
		errors.Is(err, domain.ErrRequiredConfigEmpty):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		_ = catcher.Error(op, err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "operation failed"})
	}
}
