package updater

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/UTMStack/installer/config"
)

func AuthInternalKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("internal-key") != config.GetConfig().InternalKey {
			c.JSON(401, "Unauthorized: invalid internal key")
			c.Abort()
			return
		}
	}
}
