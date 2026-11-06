package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// ACC-F11: without SetTrustedProxies, Gin trusts every hop and c.ClientIP()
// returns whatever the direct caller wrote in X-Forwarded-For. A request that
// never passed through nginx (RemoteAddr outside the Docker overlay's RFC1918
// ranges) must be read from RemoteAddr, ignoring any forged header.
func TestClientIPIgnoresForgedHeaderFromOutsideTrustedProxies(t *testing.T) {
	engine := initHTTPServer(&config{devMode: false})
	engine.GET("/ip", func(c *gin.Context) { c.String(http.StatusOK, c.ClientIP()) })

	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.RemoteAddr = "203.0.113.7:54321" // a public, non-RFC1918 address — a direct external attacker
	req.Header.Set("X-Forwarded-For", "10.10.10.10")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if got := rec.Body.String(); got != "203.0.113.7" {
		t.Fatalf("ClientIP() = %q, want the real RemoteAddr 203.0.113.7 (forged X-Forwarded-For must be ignored)", got)
	}
}

// A request actually forwarded by nginx on the overlay network must still
// resolve to the real client IP nginx appended, not an attacker-supplied one
// further up the chain.
func TestClientIPTrustsForwardedFromInternalProxy(t *testing.T) {
	engine := initHTTPServer(&config{devMode: false})
	engine.GET("/ip", func(c *gin.Context) { c.String(http.StatusOK, c.ClientIP()) })

	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.RemoteAddr = "172.18.0.5:54321" // nginx's address on the overlay network
	req.Header.Set("X-Forwarded-For", "198.51.100.9")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if got := rec.Body.String(); got != "198.51.100.9" {
		t.Fatalf("ClientIP() = %q, want 198.51.100.9 (the address nginx forwarded)", got)
	}
}
