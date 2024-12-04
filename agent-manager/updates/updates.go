package updates

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/utils"
)

var mutex sync.Mutex

type UpdatesManager struct {
	versionsCache map[string]string
}

func InitUpdatesManager() {
	updatesManager := &UpdatesManager{
		versionsCache: make(map[string]string),
	}
	go updatesManager.UpdateVersionCache()
	updatesManager.ServeDependencies()
}

func (u *UpdatesManager) UpdateVersionCache() {
	for {
		versions := map[string]string{}
		err := utils.ReadJson(config.UpdatesVersionsPath, &versions)
		if err != nil {
			utils.ALogger.ErrorF("error reading versions file: %v", err)
			time.Sleep(config.CheckEvery)
			continue
		}

		mutex.Lock()
		for c, v := range versions {
			if c == "agent-service" || c == "agent-installer" ||
				c == "collector-installer" || c == "as400-service" {
				u.versionsCache[c] = v
			}
		}
		mutex.Unlock()
		time.Sleep(config.CheckEvery)
	}
}

func (u *UpdatesManager) ServeDependencies() {
	r := gin.New()
	r.Use(
		gin.Recovery(),
		gzip.Gzip(gzip.DefaultCompression),
	)

	r.NoRoute(notFound)

	group := r.Group("/private", HTTPAuthInterceptor())
	group.StaticFS("/dependencies", http.Dir(config.UpdatesDependenciesFolder))
	group.GET("/versions", u.getVersions)

	utils.ALogger.Info("Starting HTTP server on port 8080")
	if err := r.RunTLS(":8080", config.CertPath, config.CertKeyPath); err != nil {
		utils.ALogger.ErrorF("error starting HTTP server: %v", err)
		return
	}
}

func notFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
}

func (u *UpdatesManager) getVersions(c *gin.Context) {
	mutex.Lock()
	defer mutex.Unlock()
	c.JSON(http.StatusOK, u.versionsCache)
}
