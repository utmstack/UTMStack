//go:build !windows

package service

import (
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/netblock"
)

// resolveSystemNets builds the never-block system-nets set on non-Windows hosts.
// Only server IPs are resolved here; LAN resolver/gateway discovery is a
// Windows-only path (GetAdaptersAddresses) and non-Windows is out of scope for
// Phase 1.
func resolveSystemNets(cfg config.EDRConfig) netblock.SystemNets {
	return netblock.SystemNets{ServerHosts: resolveHostIPs(cfg.Server)}
}
