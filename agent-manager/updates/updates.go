package updates

import (
	"crypto/tls"
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

	cert, err := tls.LoadX509KeyPair("/cert/utm.crt", "/cert/utm.key")
	if err != nil {
		util.Logger.ErrorF("failed to load certificates: %v", err)
	}

	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{cert},
	}

	server := &http.Server{
		Addr:      ":8080",
		Handler:   r,
		TLSConfig: tlsConfig,
	}

	util.Logger.Info("Starting HTTP server on port 8080")
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		util.Logger.ErrorF("error starting HTTP server: %v", err)
	}

}

func notFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
}
