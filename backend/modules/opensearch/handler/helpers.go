package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func writeOSError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
