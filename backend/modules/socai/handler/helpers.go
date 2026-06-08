package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
)

func writeError(c *gin.Context, msg string) {
	_ = catcher.Error(msg, nil, nil)
	c.JSON(http.StatusInternalServerError, gin.H{
		"status":  "error",
		"message": msg,
	})
}
