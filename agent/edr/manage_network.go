package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/netip"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/netblock"
	edrservice "github.com/utmstack/UTMStack/agent/edr/service"
)

type isolationController interface {
	Isolate(context.Context, netblock.IsolationPolicy) (netblock.IsolationStatus, error)
	Release(context.Context) error
	Status(context.Context) (netblock.IsolationStatus, error)
}

var newIsolationController = func() (isolationController, error) { return netblock.NewHostIsolator() }

func parseAddressList(s string) ([]netip.Addr, error) {
	var out []netip.Addr
	if s == "" {
		return out, nil
	}
	parts := strings.Split(s, ",")
	if len(parts) > 256 {
		return nil, fmt.Errorf("address list exceeds 256 entries")
	}
	for _, p := range parts {
		a, err := netip.ParseAddr(strings.TrimSpace(p))
		if err != nil || !a.IsValid() || a.IsUnspecified() || a.IsMulticast() || a.Zone() != "" {
			return nil, fmt.Errorf("expected exact IP addresses")
		}
		out = append(out, a.Unmap())
	}
	return out, nil
}
func runNetwork(args []string, out io.Writer) int {
	fail := func(err error) int { fmt.Fprintln(out, "network:", err); return 1 }
	if len(args) == 0 {
		return fail(fmt.Errorf("usage: network isolate|release|status"))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	var policy netblock.IsolationPolicy
	switch args[0] {
	case "isolate":
		flags := flag.NewFlagSet("network isolate", flag.ContinueOnError)
		flags.SetOutput(out)
		management := flags.String("management", "", "comma-separated exact management IPs, in addition to configured servers")
		resolvers := flags.String("resolver", "", "comma-separated exact DNS resolver IPs")
		allow := flags.String("allow", "", "comma-separated additional CIDRs")
		duration := flags.Duration("duration", 5*time.Minute, "finite isolation lease, at most one hour")
		if flags.Parse(args[1:]) != nil {
			return 1
		}
		if flags.NArg() != 0 {
			return fail(fmt.Errorf("unexpected isolation arguments"))
		}
		if *duration <= 0 || *duration > time.Hour {
			return fail(fmt.Errorf("isolation duration must be greater than zero and at most 1h"))
		}
		var err error
		policy.Management, err = parseAddressList(*management)
		if err != nil {
			return fail(err)
		}
		cfg, err := config.Load()
		if err != nil {
			return fail(err)
		}
		configured, err := edrservice.ManagementAddresses(ctx, cfg)
		if err != nil {
			return fail(err)
		}
		policy.Management = append(policy.Management, configured...)
		if len(policy.Management) == 0 {
			return fail(fmt.Errorf("at least one exact management IP or resolvable configured server is required"))
		}
		policy.Resolvers, err = parseAddressList(*resolvers)
		if err != nil {
			return fail(err)
		}
		discovered, resolverErr := edrservice.ResolverAddressesChecked()
		if resolverErr != nil {
			explicitUpstream := false
			for _, a := range policy.Resolvers {
				if !a.IsLoopback() {
					explicitUpstream = true
				}
			}
			if !explicitUpstream {
				return fail(resolverErr)
			}
		}
		policy.Resolvers = append(policy.Resolvers, discovered...)
		if *allow != "" {
			parts := strings.Split(*allow, ",")
			if len(parts) > 128 {
				return fail(fmt.Errorf("allow list exceeds 128 prefixes"))
			}
			for _, p := range parts {
				prefix, err := netip.ParsePrefix(strings.TrimSpace(p))
				if err != nil || prefix.Bits() == 0 {
					return fail(fmt.Errorf("allow entries must be scoped CIDRs, not a default route"))
				}
				policy.Allow = append(policy.Allow, prefix.Masked())
			}
		}
		policy.Duration = *duration
	case "release", "status":
		if len(args) != 1 {
			return fail(fmt.Errorf("unexpected arguments"))
		}
	default:
		return fail(fmt.Errorf("unknown network operation"))
	}
	controller, err := newIsolationController()
	if err != nil {
		return fail(err)
	}
	var status netblock.IsolationStatus
	switch args[0] {
	case "isolate":
		status, err = controller.Isolate(ctx, policy)
	case "release":
		err = controller.Release(ctx)
		if err == nil {
			status, err = controller.Status(ctx)
		}
	case "status":
		status, err = controller.Status(ctx)
	}
	if err != nil {
		return fail(err)
	}
	if err = json.NewEncoder(out).Encode(status); err != nil {
		return fail(err)
	}
	return 0
}
