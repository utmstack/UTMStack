//go:build linux

package service

import (
	"context"
	"errors"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/netblock"
	"github.com/utmstack/UTMStack/shared/logger"
	"io"
	"net/netip"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

func parseLinuxResolvers(b []byte) []netip.Addr {
	if len(b) > 64<<10 {
		return nil
	}
	var out []netip.Addr
	seen := map[netip.Addr]bool{}
	for _, line := range strings.Split(string(b), "\n") {
		f := strings.Fields(line)
		if len(f) < 2 || f[0] != "nameserver" {
			continue
		}
		a, err := netip.ParseAddr(f[1])
		if err != nil || a.Zone() != "" || a.IsMulticast() || a.IsUnspecified() {
			continue
		}
		a = a.Unmap()
		if !seen[a] {
			out = append(out, a)
			seen[a] = true
		}
		if len(out) == 32 {
			break
		}
	}
	return out
}
func readResolverFile(path string) ([]byte, error) {
	f, err := os.OpenFile(path, os.O_RDONLY|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("resolver configuration is not a regular file")
	}
	b, err := io.ReadAll(io.LimitReader(f, (64<<10)+1))
	if err != nil {
		return nil, err
	}
	if len(b) > 64<<10 {
		return nil, errors.New("resolver configuration exceeds 64 KiB")
	}
	return b, nil
}
func discoverLinuxResolvers(read func(string) ([]byte, error)) ([]netip.Addr, error) {
	b, err := read("/etc/resolv.conf")
	if err != nil {
		return nil, err
	}
	addresses := parseLinuxResolvers(b)
	for _, a := range addresses {
		if !a.IsLoopback() {
			return addresses, nil
		}
	}
	upstream, err := read("/run/systemd/resolve/resolv.conf")
	if err != nil {
		return addresses, errors.New("only a local DNS stub is known; supply actual upstream --resolver addresses")
	}
	seen := map[netip.Addr]bool{}
	for _, a := range addresses {
		seen[a] = true
	}
	external := false
	for _, a := range parseLinuxResolvers(upstream) {
		if a.IsLoopback() {
			continue
		}
		external = true
		if !seen[a] && len(addresses) < 32 {
			addresses = append(addresses, a)
			seen[a] = true
		}
	}
	if !external {
		return addresses, errors.New("actual upstream DNS resolver addresses unavailable")
	}
	return addresses, nil
}
func ResolverAddressesChecked() ([]netip.Addr, error) {
	return discoverLinuxResolvers(readResolverFile)
}
func ResolverAddresses() []netip.Addr {
	a, err := ResolverAddressesChecked()
	if err != nil {
		logger.Error("UTMStack EDR: resolver safety: %v", err)
	}
	return a
}
func resolveSystemNets(cfg config.EDRConfig) netblock.SystemNets {
	return resolveSystemNetsContext(context.Background(), cfg)
}

func resolveSystemNetsContext(ctx context.Context, cfg config.EDRConfig) netblock.SystemNets {
	management, err := ManagementAddresses(ctx, cfg)
	if err != nil {
		logger.Error("UTMStack EDR: network safety resolution: %v", err)
		management = nil
	}
	resolvers, err := ResolverAddressesChecked()
	if err != nil {
		logger.Error("UTMStack EDR: resolver safety: %v", err)
		management = nil
	}
	return netblock.SystemNets{ServerHosts: management, Resolvers: resolvers}
}
