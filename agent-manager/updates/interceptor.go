package updates

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/UTMStack/agent-manager/agent"
	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/utils"
)

func HTTPAuthInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		connectionKey := c.GetHeader("connection-key")
		id := c.GetHeader("id")
		key := c.GetHeader("key")
		typ := strings.ToLower(c.GetHeader("type"))

		if connectionKey == "" && id == "" && key == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "authentication is not provided"})
			return
		} else if connectionKey != "" {
			isValid := utils.IsConnectionKeyValid(fmt.Sprintf(config.PanelConnectionKeyUrl, config.GetUTMHost()), connectionKey)
			if !isValid {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid connection key"})
				return
			}
		} else if id != "" && key != "" && typ != "" {
			idInt, err := strconv.ParseUint(id, 10, 32)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "id is not valid"})
				return
			}

			switch typ {
			case "agent":
				if _, isValid := utils.IsKeyPairValid(key, uint(idInt), agent.AgentServ.CacheAgentKey); !isValid {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid key"})
					return
				}
			case "collector":
				if _, isValid := utils.IsKeyPairValid(key, uint(idInt), agent.CollectorServ.CacheCollectorKey); !isValid {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid key"})
					return
				}
			default:
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid type"})
			}
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid auth type"})
			return
		}

		c.Next()
	}
}
