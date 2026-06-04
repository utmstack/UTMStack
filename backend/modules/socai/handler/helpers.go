package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/logger"
)

func writeError(c *gin.Context, msg string) {
	logger.Error(msg)
	c.JSON(http.StatusInternalServerError, gin.H{
		"status":  "error",
		"message": msg,
	})
}
