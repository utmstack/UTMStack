package updates

import (
	"crypto/tls"
	"net/http"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
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

	group := r.Group("/private")
	group.StaticFS("/dependencies", http.Dir("/dependencies"))

	cert, err := tls.LoadX509KeyPair("/cert/utm.crt", "/cert/utm.key")
	if err != nil {
		catcher.Error("failed to load certificates", err, map[string]any{})
	}

	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
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

	catcher.Info("Starting HTTP server on port 8080", map[string]any{})
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		catcher.Error("error starting HTTP server", err, map[string]any{})
	}

}

func notFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
}
