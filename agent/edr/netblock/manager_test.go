// edr/netblock/manager_test.go
package netblock

import (
	"net/netip"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
)

type memSpool struct {
	mu    sync.Mutex
	lines []string
}

func (m *memSpool) Append(s string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lines = append(m.lines, s)
	return nil
}

func TestManagerBlocksAndEmits(t *testing.T) {
	c, _ := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	defer c.Close()
	cfg := config.Default()
	cfg.Blocklist.AllowPrivateRanges = false
	sp := &memSpool{}
	m := NewManager(Deps{Cfg: cfg, Cache: c, Spool: sp})

	// Directly drive the desired-set application (no live feed in the test).
	// A public, non-allowlisted indicator survives filtering (nothing dropped).
	// ActiveFilters reflects the REAL OS blocker count, which is 0 for the host's
	// no-op blocker — so we assert on the allowlist outcome, not the filter count.
	i, _ := ParseIndicator("9.9.9.9", "ip", 1)
	m.applyIndicators([]Indicator{i})
	if m.Health().AllowlistedDropped != 0 {
		t.Fatalf("allowlisted-dropped = %d, want 0 (9.9.9.9 is not allowlisted)", m.Health().AllowlistedDropped)
	}
	// On host, WFP init returns the no-op blocker without error → wfpOK true →
	// enforcement is not degraded (status must not claim degraded protection).
	if m.Health().EnforcementDegraded {
		t.Fatal("enforcement must not be degraded on host (no-op blocker, wfpOK true)")
	}

	// A drop to a BLOCKLISTED IP → a branded network_watcher event + cache record.
	// The store must hold the indicator (the WFP net-event feed delivers all system
	// drops; onBlockEvent emits only for store matches).
	m.store.Swap([]Indicator{i})
	m.onBlockEvent(BlockEvent{RemoteIP: netip.MustParseAddr("9.9.9.9"), RemotePort: 443, Direction: "outbound"})
	if len(sp.lines) != 1 {
		t.Fatalf("want 1 spooled event, got %d", len(sp.lines))
	}
	// A drop to a NON-blocklisted IP must NOT emit (system-wide drops are filtered).
	m.onBlockEvent(BlockEvent{RemoteIP: netip.MustParseAddr("8.8.8.8"), RemotePort: 443, Direction: "outbound"})
	if len(sp.lines) != 1 {
		t.Fatalf("non-blocklisted drop must not emit; want 1 line, got %d", len(sp.lines))
	}

	// Allowlisted target must be dropped from the desired set.
	cfg2 := config.Default()
	cfg2.Blocklist.AllowPrivateRanges = true
	m2 := NewManager(Deps{Cfg: cfg2, Cache: c, Spool: &memSpool{}})
	priv, _ := ParseIndicator("10.1.2.3", "ip", 1)
	m2.applyIndicators([]Indicator{priv})
	if m2.Health().AllowlistedDropped != 1 {
		t.Fatalf("allowlisted-dropped = %d, want 1", m2.Health().AllowlistedDropped)
	}
	_ = time.Second
}

func TestManagerDNSReactiveBlock(t *testing.T) {
	c, _ := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	defer c.Close()
	cfg := config.Default()
	sp := &memSpool{}
	m := NewManager(Deps{Cfg: cfg, Cache: c, Spool: sp})
	// Blocklist holds a domain.
	d, _ := ParseIndicator("evil.com", "domain", 2)
	m.store.Swap([]Indicator{d})

	// A resolution of a subdomain to two IPs → event naming the domain.
	m.onDNSEvent(DNSEvent{
		Name: "c2.evil.com",
		IPs:  []netip.Addr{netip.MustParseAddr("9.9.9.9"), netip.MustParseAddr("9.9.9.10")},
		PID:  1234,
	})
	if len(sp.lines) != 1 {
		t.Fatalf("want 1 domain event, got %d", len(sp.lines))
	}
	if !strings.Contains(sp.lines[0], `"domain":"c2.evil.com"`) ||
		!strings.Contains(sp.lines[0], `"indicator":"evil.com"`) {
		t.Fatalf("event missing domain/indicator: %s", sp.lines[0])
	}
	// A resolution of a NON-blocklisted name → no event, no block.
	m.onDNSEvent(DNSEvent{Name: "good.com", IPs: []netip.Addr{netip.MustParseAddr("8.8.8.8")}, PID: 1})
	if len(sp.lines) != 1 {
		t.Fatalf("non-blocklisted resolution must not emit; got %d", len(sp.lines))
	}
}
