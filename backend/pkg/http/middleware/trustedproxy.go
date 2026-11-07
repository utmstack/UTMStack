package middleware

import (
	"context"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
)

const clientIPContextKey = "resolved_client_ip"

// trustedProxyIPs holds the current set of trusted-proxy IPs, swapped
// atomically by WatchTrustedProxy. Gin's own SetTrustedProxies mechanism
// mutates an unsynchronized field on *gin.Engine that ClientIP() reads with no
// locking — updating it from a background refresh goroutine while requests
// are concurrently being served is a real, race-detector-confirmed data race.
// ResolveClientIP/ClientIP below read this instead, so the proxy's resolved
// IP(s) can be refreshed at runtime (ACC-F11) without ever touching gin.Engine
// state after startup.
var trustedProxyIPs atomic.Pointer[map[string]struct{}]

func init() {
	empty := map[string]struct{}{}
	trustedProxyIPs.Store(&empty)
}

// SetTrustedProxyIPsForTest overrides the trusted-proxy set. Exported only for
// tests outside this package that exercise ResolveClientIP/ClientIP directly.
func SetTrustedProxyIPsForTest(ips ...string) { setTrustedProxyIPs(ips) }

func setTrustedProxyIPs(ips []string) {
	m := make(map[string]struct{}, len(ips))
	for _, ip := range ips {
		m[ip] = struct{}{}
	}
	trustedProxyIPs.Store(&m)
}

// ResolveClientIP computes the real client IP once per request and stashes it
// on the gin.Context for ClientIP to read back. A request is only trusted to
// name its own "real" IP via X-Forwarded-For when it came from an address in
// trustedProxyIPs — resolved to the "frontend" Swarm service's own nginx by
// WatchTrustedProxy, never a CIDR (ACC-F11: a CIDR wide enough to survive a
// redeploy's dynamic overlay IP also covers every other container on the
// same network, and in a test environment the tester's own machine). nginx
// appends the address it saw to the end of X-Forwarded-For
// ($proxy_add_x_forwarded_for, not a plain overwrite), so the last entry is
// the one hop actually trusted here.
func ResolveClientIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(clientIPContextKey, resolveClientIP(c.Request.RemoteAddr, c.Request.Header.Get("X-Forwarded-For")))
		c.Next()
	}
}

func resolveClientIP(remoteAddr, xff string) string {
	host := remoteAddr
	if h, _, err := net.SplitHostPort(remoteAddr); err == nil {
		host = h
	}
	set := *trustedProxyIPs.Load()
	if _, trusted := set[host]; !trusted || xff == "" {
		return host
	}
	parts := strings.Split(xff, ",")
	return strings.TrimSpace(parts[len(parts)-1])
}

// ClientIP returns the client IP resolved by ResolveClientIP. Every call site
// that used to read c.ClientIP() directly must use this instead (ACC-F11) —
// see the package doc on trustedProxyIPs for why.
func ClientIP(c *gin.Context) string {
	if v, ok := c.Get(clientIPContextKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return c.ClientIP()
}

// WatchTrustedProxy keeps the trusted-proxy set pinned to exactly the reverse
// proxy that sits in front of this backend — the "frontend" Swarm service's
// own nginx. Docker's embedded DNS resolves "tasks.<service>" to the live
// container IP(s) of a Swarm service — the same mechanism nginx's own
// `resolver 127.0.0.11` directive already relies on. This blocks until that
// name resolves at least once (the frontend container starts only once the
// backend's health check passes, so it may not exist for the first few
// seconds), then refreshes on an interval so a rescheduled frontend task's
// new IP is picked up without a backend restart. Until the first successful
// resolution, nothing is trusted — the safe default: no X-Forwarded-For is
// honoured and ClientIP returns the raw RemoteAddr.
func WatchTrustedProxy(ctx context.Context, host string, interval time.Duration) {
	watchTrustedProxy(ctx, host, interval, func(host string) []string {
		addrs, err := net.LookupHost("tasks." + host)
		if err != nil {
			return nil
		}
		return addrs
	})
}

// watchTrustedProxy takes the DNS lookup as a parameter so tests can inject a
// fake resolver instead of depending on real Docker DNS.
func watchTrustedProxy(ctx context.Context, host string, interval time.Duration, resolve func(host string) []string) {
	apply := func() bool {
		ips := resolve(host)
		if len(ips) == 0 {
			return false
		}
		setTrustedProxyIPs(ips)
		return true
	}

	for !apply() {
		select {
		case <-time.After(2 * time.Second):
		case <-ctx.Done():
			return
		}
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if !apply() {
				catcher.Warn("could not re-resolve trusted proxy; keeping the last known address(es)", map[string]any{"host": host})
			}
		case <-ctx.Done():
			return
		}
	}
}
