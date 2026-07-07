// edr/netblock/manager_test.go
package netblock

import (
	"net/netip"
	"path/filepath"
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
	i, _ := ParseIndicator("9.9.9.9", "ip", 1)
	m.applyIndicators([]Indicator{i})
	if m.Health().ActiveFilters != 1 {
		t.Fatalf("active filters = %d, want 1 (noop blocker counts via enforcer)", m.Health().ActiveFilters)
	}

	// A drop audit event → a branded network_watcher event + a cache record.
	m.onBlockEvent(BlockEvent{RemoteIP: netip.MustParseAddr("9.9.9.9"), RemotePort: 443, Direction: "outbound"})
	if len(sp.lines) != 1 {
		t.Fatalf("want 1 spooled event, got %d", len(sp.lines))
	}

	// Allowlisted target must be dropped from the desired set.
	cfg2 := config.Default()
	cfg2.Blocklist.AllowPrivateRanges = true
	m2 := NewManager(Deps{Cfg: cfg2, Cache: c, Spool: &memSpool{}})
	priv, _ := ParseIndicator("10.1.2.3", "ip", 1)
	m2.applyIndicators([]Indicator{priv})
	if m2.Health().ActiveFilters != 0 {
		t.Fatal("allowlisted indicator must not become a filter")
	}
	if m2.Health().AllowlistedDropped != 1 {
		t.Fatalf("allowlisted-dropped = %d, want 1", m2.Health().AllowlistedDropped)
	}
	_ = time.Second
}
