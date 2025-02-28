package updates

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/utils"
)

func InitUpdatesManager() {
	ServeDependencies()
}

func ServeDependencies() {
	r := gin.New()
	r.Use(
		gin.Recovery(),
		gzip.Gzip(gzip.DefaultCompression),
	)

	r.NoRoute(notFound)

	group := r.Group("/private", HTTPAuthInterceptor())
	group.StaticFS("/dependencies", http.Dir(config.UpdatesDependenciesFolder))
	group.GET("/version", getVersion)

	utils.ALogger.Info("Starting HTTP server on port 8080")
	if err := r.RunTLS(":8080", config.CertPath, config.CertKeyPath); err != nil {
		utils.ALogger.ErrorF("error starting HTTP server: %v", err)
		return
	}
}

func notFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
}

func getVersion(c *gin.Context) {
	release := models.Release{}
	err := utils.ReadJson(config.UpdatesVersionsPath, &release)
	if err != nil {
		utils.ALogger.ErrorF("error reading versions file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error reading versions file"})
		return
	}

	version := models.Version{
		Version: fmt.Sprintf("%s-%s", release.Version, release.Edition),
	}

	c.JSON(http.StatusOK, version)
}
