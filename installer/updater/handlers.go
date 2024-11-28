package updater

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	twsdk "github.com/threatwinds/go-sdk"
	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/utils"
)

func notFound(c *gin.Context) {
	requestedUrl := c.Request.URL.Path
	twsdk.Logger().ErrorF("404 Not Found: %s", requestedUrl)
	e := twsdk.Logger().ErrorF(fmt.Sprintf("404 Not Found: IP %s requested %s", c.ClientIP(), requestedUrl))
	e.GinError(c)
}

func getEdition(c *gin.Context) {
	tlsConfig, err := utils.LoadTLSCredentials(config.CertFilePath)
	if err != nil {
		utils.Logger.ErrorF("error loading tls credentials: %v", err)
		c.JSON(500, gin.H{"error": "error loading tls credentials"})
		return
	}

	resp, status, err := utils.DoReq[string](
		GetUpdaterClient().Server+config.GetEditionEndpoint,
		nil,
		http.MethodGet, map[string]string{"instance-id": GetUpdaterClient().InstanceID, "instance-key": GetUpdaterClient().InstanceKey},
		tlsConfig,
	)
	if err != nil {
		utils.Logger.ErrorF("error getting edition: %v", err)
		c.JSON(500, fmt.Sprintf("error getting edition: %v", err))
		return
	} else if status != http.StatusOK {
		utils.Logger.ErrorF("error getting edition: %v", err)
		c.JSON(500, fmt.Sprintf("error getting edition: %v", err))
		return
	}

	c.JSON(200, resp)
}

func changeUpdateWindow(c *gin.Context) {
	var window config.MaintenanceWindow
	if err := c.BindJSON(&window); err != nil {
		utils.Logger.ErrorF("error binding json: %v", err)
		c.JSON(500, gin.H{"error": "error binding json"})
		return
	}

	if err := config.SaveWindowsConfig(window); err != nil {
		utils.Logger.ErrorF("error saving window config: %v", err)
		c.JSON(500, gin.H{"error": "error saving window config"})
		return
	}

	c.JSON(200, gin.H{"message": "window config saved"})
}

func getVersions(c *gin.Context) {
	versions := GetVersions()
	c.JSON(200, versions)
}
