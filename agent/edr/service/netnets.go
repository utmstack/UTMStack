package service

import (
	"net"
	"net/netip"
	"strings"
)

// resolveHostIPs parses the host out of the configured server string (which may
// be a URL like "https://host:port" or a bare "host"/"host:port") and resolves
// it to canonical netip.Addr values. These become never-block ServerHosts in the
// blocklist allowlist so the EDR can never sever the host's link to the platform.
// Returns nil on empty input or any lookup error — never panics.
func resolveHostIPs(server string) []netip.Addr {
	host := hostFromServer(server)
	if host == "" {
		return nil
	}
	// If the host is already a literal IP, use it directly (no DNS).
	if a, err := netip.ParseAddr(host); err == nil {
		return []netip.Addr{a.Unmap()}
	}
	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return nil
	}
	out := make([]netip.Addr, 0, len(ips))
	for _, ip := range ips {
		var a netip.Addr
		var ok bool
		if v4 := ip.To4(); v4 != nil {
			a, ok = netip.AddrFromSlice(v4)
		} else if v16 := ip.To16(); v16 != nil {
			a, ok = netip.AddrFromSlice(v16)
		}
		if ok {
			out = append(out, a.Unmap())
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// hostFromServer strips an optional scheme ("https://") and an optional ":port"
// suffix from a server string, returning just the hostname/IP. It tolerates bare
// hosts, "host:port", and full URLs.
func hostFromServer(server string) string {
	s := strings.TrimSpace(server)
	if s == "" {
		return ""
	}
	// Strip scheme.
	if i := strings.Index(s, "://"); i >= 0 {
		s = s[i+3:]
	}
	// Drop any path/query.
	if i := strings.IndexAny(s, "/?"); i >= 0 {
		s = s[:i]
	}
	// Bracketed IPv6 literal: [::1]:9000 or [::1].
	if strings.HasPrefix(s, "[") {
		if j := strings.Index(s, "]"); j >= 0 {
			return s[1:j]
		}
		return strings.TrimPrefix(s, "[")
	}
	// Strip a trailing :port (only for a single colon — a bare IPv6 literal has
	// multiple colons and no port here since it would be bracketed above).
	if strings.Count(s, ":") == 1 {
		if i := strings.LastIndex(s, ":"); i >= 0 {
			s = s[:i]
		}
	}
	return s
}
