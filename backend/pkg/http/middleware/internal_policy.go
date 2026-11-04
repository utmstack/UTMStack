package middleware

import (
	"net"
	"net/http"
	"strings"
)

// InternalPolicy narrows where the shared internal key is honoured. The key is
// one secret held by every service in the stack, so on its own it is a master
// key to the whole API; the policy makes it good only for the machine-to-machine
// routes those services actually call, and optionally only from their network.
// A request carrying the key anywhere else is treated as unauthenticated.
type InternalPolicy struct {
	// Paths are matched on whole path segments: "/api/v1/mcp" covers
	// "/api/v1/mcp" and "/api/v1/mcp/x" but not "/api/v1/mcpx".
	Paths []string
	// CIDRs, when set, restricts the callers by remote address.
	CIDRs []*net.IPNet
}

func (p InternalPolicy) permitsPath(path string) bool {
	for _, allowed := range p.Paths {
		if path == allowed || strings.HasPrefix(path, strings.TrimRight(allowed, "/")+"/") {
			return true
		}
	}
	return false
}

func (p InternalPolicy) permitsAddr(ip net.IP) bool {
	if len(p.CIDRs) == 0 {
		return true
	}
	for _, n := range p.CIDRs {
		if ip != nil && n.Contains(ip) {
			return true
		}
	}
	return false
}

// Permits reports whether the request may use the internal key, and if not, why.
func (p InternalPolicy) Permits(r *http.Request, clientIP string) (bool, string) {
	if !p.permitsPath(r.URL.Path) {
		return false, "path is not on the internal allowlist"
	}
	if !p.permitsAddr(net.ParseIP(clientIP)) {
		return false, "caller address is outside the allowed networks"
	}
	return true, ""
}

// ParseCIDRs reads a comma-separated list such as "10.0.0.0/8,172.16.0.0/12".
func ParseCIDRs(list string) ([]*net.IPNet, error) {
	var out []*net.IPNet
	for _, part := range strings.Split(list, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		_, n, err := net.ParseCIDR(part)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}
