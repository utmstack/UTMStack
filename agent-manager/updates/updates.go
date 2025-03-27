package updates

import (
	"net/http"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/utmstack/UTMStack/agent-manager/auth"
	"github.com/utmstack/UTMStack/agent-manager/util"
)

type Version struct {
	Version string `json:"version"`
}

func InitUpdatesManager() {
	go ServeDependencies()
}

func ServeDependencies() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(
		gin.Recovery(),
		gzip.Gzip(gzip.DefaultCompression),
	)

	r.NoRoute(notFound)

	group := r.Group("/private", auth.HTTPAuthInterceptor())
	group.StaticFS("/dependencies", http.Dir("/dependencies"))

	util.Logger.Info("Starting HTTP server on port 8080")
	if err := r.RunTLS(":8080", "/cert/utm.crt", "/cert/utm.key"); err != nil {
		util.Logger.ErrorF("error starting HTTP server: %v", err)
		return
	}
}

func notFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
}
