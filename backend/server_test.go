package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

// ACC-F11: middleware.ClientIP must only honour X-Forwarded-For from a
// request that came through a trusted proxy — never from a direct caller.
// The resolution/refresh of which IP(s) count as trusted (pinned to the
// "frontend" Swarm service, not a CIDR) is covered separately in
// pkg/http/middleware/trustedproxy_test.go, since it depends on Docker DNS
// that isn't available here.
func newClientIPEngine(trustedIP string) *gin.Engine {
	engine := gin.New()
	engine.Use(middleware.ResolveClientIP())
	engine.GET("/ip", func(c *gin.Context) { c.String(http.StatusOK, middleware.ClientIP(c)) })
	middleware.SetTrustedProxyIPsForTest(trustedIP)
	return engine
}

func TestClientIPIgnoresForgedHeaderFromOutsideTrustedProxies(t *testing.T) {
	engine := newClientIPEngine("172.18.0.5")

	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.RemoteAddr = "203.0.113.7:54321" // a public, non-RFC1918 address — a direct external attacker
	req.Header.Set("X-Forwarded-For", "10.10.10.10")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if got := rec.Body.String(); got != "203.0.113.7" {
		t.Fatalf("ClientIP() = %q, want the real RemoteAddr 203.0.113.7 (forged X-Forwarded-For must be ignored)", got)
	}
}

// A request actually forwarded by the trusted proxy must still resolve to the
// real client IP it appended, not an attacker-supplied one further up the
// chain.
func TestClientIPTrustsForwardedFromInternalProxy(t *testing.T) {
	engine := newClientIPEngine("172.18.0.5")

	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.RemoteAddr = "172.18.0.5:54321" // the trusted proxy's address
	req.Header.Set("X-Forwarded-For", "198.51.100.9")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if got := rec.Body.String(); got != "198.51.100.9" {
		t.Fatalf("ClientIP() = %q, want 198.51.100.9 (the address the proxy forwarded)", got)
	}
}
