package auth

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
)

func HTTPAuthInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		connectionKey := c.GetHeader("connection-key")
		id := c.GetHeader("id")
		key := c.GetHeader("key")
		requestURL := c.Request.URL.Path

		if connectionKey == "" && id == "" && key == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "authentication is not provided"})
			return
		} else if connectionKey != "" {
			if err := authenticateRequest(metadata.New(map[string]string{"connection-key": connectionKey}), "connection-key"); err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid connection key"})
				return
			}
		} else if id != "" && key != "" {
			idInt, err := strconv.ParseUint(id, 10, 32)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "id is not valid"})
				return
			}

			if err := checkKeyAuth(key, idInt, requestURL); err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid key"})
				return
			}

		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid auth type"})
			return
		}

		c.Next()
	}
}
