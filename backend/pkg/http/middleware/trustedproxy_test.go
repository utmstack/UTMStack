package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

func resolvedClientIP(t *testing.T, remoteAddr, xff string) string {
	t.Helper()
	engine := gin.New()
	engine.Use(ResolveClientIP())
	engine.GET("/ip", func(c *gin.Context) { c.String(http.StatusOK, ClientIP(c)) })

	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.RemoteAddr = remoteAddr
	if xff != "" {
		req.Header.Set("X-Forwarded-For", xff)
	}
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec.Body.String()
}

// Until the proxy resolves for the first time, nothing is trusted: a forged
// header from any address, including one inside what used to be the broad
// RFC1918 allow-list, must be ignored.
func TestWatchTrustedProxyStartsWithNothingTrusted(t *testing.T) {
	setTrustedProxyIPs(nil) // reset shared state: tests in this package run in one process

	if got := resolvedClientIP(t, "10.11.11.50:1234", "203.0.113.7"); got != "10.11.11.50" {
		t.Fatalf("ClientIP() = %q, want the real RemoteAddr (no proxy resolved yet)", got)
	}
}

// ACC-F11: once the proxy resolves to a specific IP, only a request whose
// RemoteAddr is exactly that IP gets its X-Forwarded-For honoured — a
// different address on the same private network (what a CIDR-based trust
// list would have let through, and what the live bypass on dev-19 exploited)
// must not be.
func TestWatchTrustedProxyPinsToResolvedIPNotACIDR(t *testing.T) {
	setTrustedProxyIPs(nil)

	var calls atomic.Int32
	resolve := func(host string) []string {
		calls.Add(1)
		if host != "frontend" {
			t.Errorf("resolve called with host %q, want %q", host, "frontend")
		}
		return []string{"172.18.0.5"}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go watchTrustedProxy(ctx, "frontend", time.Hour, resolve)

	deadline := time.Now().Add(2 * time.Second)
	for calls.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}

	if got := resolvedClientIP(t, "172.18.0.5:1234", "198.51.100.9"); got != "198.51.100.9" {
		t.Fatalf("ClientIP() from the resolved proxy = %q, want 198.51.100.9", got)
	}
	// Same private network, different host than the resolved proxy — must not
	// be trusted, unlike the old 10.0.0.0/8-style CIDR fix.
	if got := resolvedClientIP(t, "10.11.11.50:1234", "203.0.113.7"); got != "10.11.11.50" {
		t.Fatalf("ClientIP() from a non-proxy host on the same private network = %q, want the real RemoteAddr", got)
	}
}

// A refresh that finds nothing (e.g. a transient DNS hiccup) must not blow
// away a previously-known-good trusted set.
func TestWatchTrustedProxyKeepsLastKnownGoodOnFailedRefresh(t *testing.T) {
	setTrustedProxyIPs(nil)

	var fail atomic.Bool
	resolve := func(string) []string {
		if fail.Load() {
			return nil
		}
		return []string{"172.18.0.5"}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go watchTrustedProxy(ctx, "frontend", 10*time.Millisecond, resolve)

	deadline := time.Now().Add(2 * time.Second)
	for resolvedClientIP(t, "172.18.0.5:1234", "198.51.100.9") != "198.51.100.9" && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}

	fail.Store(true)
	time.Sleep(50 * time.Millisecond) // let a few failed refresh ticks pass

	if got := resolvedClientIP(t, "172.18.0.5:1234", "198.51.100.9"); got != "198.51.100.9" {
		t.Fatalf("ClientIP() after a failed refresh = %q, want the last known good 198.51.100.9", got)
	}
}
