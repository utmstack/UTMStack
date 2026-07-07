// edr/netblock/allowlist.go
package netblock

import (
	"net/netip"
	"strings"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

// SystemNets are host-derived addresses that must never be blocked (resolved by
// the caller at startup so this stays pure/testable).
type SystemNets struct {
	Resolvers   []netip.Addr
	Gateways    []netip.Addr
	ServerHosts []netip.Addr // UTMStack platform / mirror IPs
}

// Allowlist is the never-block set: built-in critical ranges, optional private
// ranges, host system nets, and admin-configured IP/CIDR entries.
type Allowlist struct {
	exact    map[netip.Addr]struct{}
	prefixes []netip.Prefix
}

var builtinAlways = []string{
	"127.0.0.0/8", "::1/128", // loopback
	"169.254.0.0/16", "fe80::/10", // link-local
	"224.0.0.0/4", "ff00::/8", // multicast
	"255.255.255.255/32",
}

var privateRanges = []string{
	"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "fc00::/7",
}

func BuildAllowlist(cfg config.EDRConfig, sys SystemNets) *Allowlist {
	al := &Allowlist{exact: map[netip.Addr]struct{}{}}
	addPfx := func(list []string) {
		for _, s := range list {
			if p, err := netip.ParsePrefix(s); err == nil {
				al.prefixes = append(al.prefixes, p.Masked())
			}
		}
	}
	addPfx(builtinAlways)
	if cfg.Blocklist.AllowPrivateRanges {
		addPfx(privateRanges)
	}
	// Normalize with Unmap() so an IPv4 address and its 4in6 form share one key,
	// consistent with the matcher (Task 4).
	for _, a := range sys.Resolvers {
		al.exact[a.Unmap()] = struct{}{}
	}
	for _, a := range sys.Gateways {
		al.exact[a.Unmap()] = struct{}{}
	}
	for _, a := range sys.ServerHosts {
		al.exact[a.Unmap()] = struct{}{}
	}
	// Admin-configured network exemptions live in the unified allowlist (Networks).
	for _, s := range cfg.Allowlist.Networks {
		s = strings.TrimSpace(s)
		if a, err := netip.ParseAddr(s); err == nil {
			al.exact[a.Unmap()] = struct{}{}
			continue
		}
		if p, err := netip.ParsePrefix(s); err == nil {
			al.prefixes = append(al.prefixes, p.Masked())
		}
	}
	return al
}

func (a *Allowlist) Allowed(addr netip.Addr) bool {
	addr = addr.Unmap()
	if _, ok := a.exact[addr]; ok {
		return true
	}
	for _, p := range a.prefixes {
		if p.Contains(addr) {
			return true
		}
	}
	return false
}
