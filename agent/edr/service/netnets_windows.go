//go:build windows

package service

import (
	"net/netip"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/netblock"
)

// resolveSystemNets builds the never-block system-nets set on Windows: the
// platform server IPs plus (in a later plan) the host's DNS resolvers and
// default gateways, so blocklist enforcement can never cut the host off from
// name resolution or its route to the platform.
func resolveSystemNets(cfg config.EDRConfig) netblock.SystemNets {
	return netblock.SystemNets{
		ServerHosts: resolveHostIPs(cfg.Server),
		Resolvers:   windowsDNSServers(),
		Gateways:    windowsGateways(),
	}
}

// windowsDNSServers returns the host's configured DNS resolver addresses.
// Plan 1: LAN DNS/gateway are RFC1918 and covered by AllowPrivateRanges
// (default true); public-resolver allowlisting via GetAdaptersAddresses is a
// follow-up.
func windowsDNSServers() []netip.Addr { return nil }

// windowsGateways returns the host's default gateway addresses.
// Plan 1: LAN DNS/gateway are RFC1918 and covered by AllowPrivateRanges
// (default true); public-resolver allowlisting via GetAdaptersAddresses is a
// follow-up.
func windowsGateways() []netip.Addr { return nil }
