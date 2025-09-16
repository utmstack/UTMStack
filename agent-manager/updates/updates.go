package updates

import (
	"crypto/tls"
	"net/http"

	"github.com/gin-contrib/gzip"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/utils"
)

func InitUpdatesManager() {
	ServeDependencies()
}

func ServeDependencies() {
	utils.ALogger.LogF(100, "Serving dependencies from %s", config.UpdatesDependenciesFolder)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(
		gin.Recovery(),
		gzip.Gzip(gzip.DefaultCompression),
	)

	r.NoRoute(notFound)

	group := r.Group("/private")
	group.StaticFS("/dependencies", http.Dir(config.UpdatesDependenciesFolder))

	loadedCert, err := tls.LoadX509KeyPair(config.CertPath, config.CertKeyPath)
	if err != nil {
		utils.ALogger.Fatal("failed to load TLS credentials: %v", err)
	}

	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{loadedCert},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},

		PreferServerCipherSuites: true,
	}

	server := &http.Server{
		Addr:      ":8080",
		Handler:   r,
		TLSConfig: tlsConfig,
	}

	utils.ALogger.Info("Starting HTTP server on port 8080")
	if err := server.ListenAndServeTLS("", ""); err != nil {
		utils.ALogger.ErrorF("error starting HTTP server: %v", err)
		return
	}
}

func notFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
}
