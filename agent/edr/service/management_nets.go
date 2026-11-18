package service

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

// ManagementAddresses protects every configured management/update endpoint.
// Any unresolved configured host refuses the safety policy rather than silently
// dropping that endpoint from the never-block set.
func ManagementAddresses(ctx context.Context, cfg config.EDRConfig) ([]netip.Addr, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	endpoints := []string{cfg.Server, cfg.Blocklist.MirrorBaseURL}
	if strings.Contains(cfg.SigMirror, "://") {
		endpoints = append(endpoints, cfg.SigMirror)
	}
	seen := map[netip.Addr]bool{}
	var out []netip.Addr
	for _, endpoint := range endpoints {
		if endpoint == "" {
			continue
		}
		if len(endpoint) > 4096 {
			return nil, fmt.Errorf("management endpoint exceeds length limit")
		}
		host := endpoint
		if strings.Contains(endpoint, "://") {
			u, err := url.Parse(endpoint)
			if err != nil || u.Hostname() == "" {
				return nil, fmt.Errorf("invalid management endpoint")
			}
			host = u.Hostname()
		} else if h, _, err := net.SplitHostPort(endpoint); err == nil {
			host = h
		}
		host = strings.Trim(host, "[]")
		var addrs []netip.Addr
		if a, err := netip.ParseAddr(host); err == nil {
			addrs = []netip.Addr{a}
		} else {
			var err error
			addrs, err = net.DefaultResolver.LookupNetIP(ctx, "ip", host)
			if err != nil {
				return nil, fmt.Errorf("cannot resolve management endpoint %q: %w", host, err)
			}
		}
		if len(addrs) == 0 {
			return nil, fmt.Errorf("management endpoint has no addresses")
		}
		for _, a := range addrs {
			a = a.Unmap()
			if !a.IsValid() || a.IsUnspecified() || a.IsMulticast() {
				return nil, fmt.Errorf("invalid management address")
			}
			if !seen[a] {
				if len(out) >= 256 {
					return nil, fmt.Errorf("management address limit exceeded")
				}
				seen[a] = true
				out = append(out, a)
			}
		}
	}
	return out, nil
}
