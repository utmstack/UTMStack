package updates

import (
	"crypto/tls"
	"net/http"
	"os"
	"time"

	"github.com/threatwinds/go-sdk/catcher"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/UTMStack/agent-manager/config"
)

func InitUpdatesManager() {
	ServeDependencies()
}

func ServeDependencies() {
	catcher.Info("Serving dependencies", map[string]any{"path": config.UpdatesDependenciesFolder, "process": "agent-manager"})

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(
		gin.Recovery(),
	)

	r.NoRoute(notFound)

	group := r.Group("/private")
	group.StaticFS("/dependencies", http.Dir(config.UpdatesDependenciesFolder))

	loadedCert, err := tls.LoadX509KeyPair(config.CertPath, config.CertKeyPath)
	if err != nil {
		_ = catcher.Error("failed to load TLS credentials", err, map[string]any{"process": "agent-manager"})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{loadedCert},
		MinVersion:   tls.VersionTLS12,
		MaxVersion:   tls.VersionTLS13,
		CipherSuites: []uint16{
			// TLS 1.2 secure cipher suites - RSA key exchange
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			// TLS 1.2 secure cipher suites - ECDSA key exchange (for ECDSA certificates)
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
		CurvePreferences: []tls.CurveID{
			tls.X25519,    // Modern and fast
			tls.CurveP256, // NIST P-256
			tls.CurveP384, // NIST P-384
			tls.CurveP521, // NIST P-521
		},
	}

	server := &http.Server{
		Addr:      ":8080",
		Handler:   r,
		TLSConfig: tlsConfig,
	}

	catcher.Info("Starting HTTP server on port 8080", map[string]any{"process": "agent-manager"})
	if err := server.ListenAndServeTLS("", ""); err != nil {
		_ = catcher.Error("error starting HTTP server", err, map[string]any{"process": "agent-manager"})
		return
	}
}

func notFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
}
