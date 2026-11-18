package netblock

import (
	"errors"
	"net/netip"
	"path/filepath"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
)

type failingBlocker struct{ *fakeBlocker }

func (f failingBlocker) AddIP(netip.Addr, string) error { return errors.New("injected filter failure") }

func TestDNSNeverClaimsBlockingOnBackendFailureOrAllowlistedAddress(t *testing.T) {
	for _, kind := range []string{"initialization", "filter", "allowlisted", "partial"} {
		t.Run(kind, func(t *testing.T) {
			c, err := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			cfg := config.Default()
			cfg.Allowlist.Networks = []string{"192.0.2.1/32"}
			sp := &memSpool{}
			fb := newFake()
			factory := func(string) (Blocker, error) {
				if kind == "initialization" {
					return nil, errors.New("unavailable")
				}
				if kind == "filter" {
					return failingBlocker{fb}, nil
				}
				return fb, nil
			}
			m := NewManager(Deps{Cfg: cfg, Cache: c, Spool: sp, NewBlocker: factory})
			ind, _ := ParseIndicator("test.invalid", "domain", 1)
			m.store.Swap([]Indicator{ind})
			ips := []netip.Addr{netip.MustParseAddr("192.0.2.2")}
			if kind == "allowlisted" {
				ips = []netip.Addr{netip.MustParseAddr("192.0.2.1")}
			}
			if kind == "partial" {
				ips = append(ips, netip.MustParseAddr("192.0.2.1"))
			}
			m.onDNSEvent(DNSEvent{Name: "test.invalid", IPs: ips})
			if len(sp.lines) != 1 || !strings.Contains(sp.lines[0], `"action":"detected"`) {
				t.Fatalf("unearned blocking claim: %v", sp.lines)
			}
			if _, blocked := fb.set[netip.MustParseAddr("192.0.2.1")]; blocked {
				t.Fatal("allowlisted IP blocked")
			}
			if kind == "initialization" && !m.Health().EnforcementDegraded {
				t.Fatal("failed backend reported healthy")
			}
		})
	}
}
