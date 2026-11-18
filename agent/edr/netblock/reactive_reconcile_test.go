package netblock

import (
	"errors"
	"net/netip"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
)

type retryBlocker struct {
	*fakeBlocker
	failRemove, failReset bool
}

func (b *retryBlocker) RemoveIP(a netip.Addr) error {
	if b.failRemove {
		return errors.New("injected remove failure")
	}
	return b.fakeBlocker.RemoveIP(a)
}
func (b *retryBlocker) Reset() error {
	if b.failReset {
		return errors.New("injected reset failure")
	}
	return b.fakeBlocker.Reset()
}

func reactiveManager(t *testing.T, b Blocker) (*Manager, config.EDRConfig, *memSpool) {
	t.Helper()
	c, err := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	cfg := config.Default()
	sp := &memSpool{}
	m := NewManager(Deps{Cfg: cfg, Cache: c, Spool: sp, NewBlocker: func(string) (Blocker, error) { return b, nil }})
	domain, _ := ParseIndicator("evil.example", "domain", 1)
	m.store.Swap([]Indicator{domain})
	return m, cfg, sp
}

func TestReactiveAllowlistReloadRemovalAndRetry(t *testing.T) {
	b := &retryBlocker{fakeBlocker: newFake()}
	m, cfg, _ := reactiveManager(t, b)
	a := netip.MustParseAddr("192.0.2.5")
	m.onDNSEvent(DNSEvent{Name: "evil.example", IPs: []netip.Addr{a}})
	if b.Count() != 1 {
		t.Fatal("precondition: reactive filter absent")
	}
	// Raw-feed membership must not imply ownership when the address is exempt.
	ip, _ := ParseIndicator(a.String(), "ip", 1)
	m.store.Swap([]Indicator{ip})
	b.failRemove = true
	cfg.Allowlist.Networks = []string{a.String()}
	m.Reload(cfg, SystemNets{})
	if b.Count() != 1 || m.reactive.Active() != 1 || !m.Health().EnforcementDegraded {
		t.Fatal("failed exempt-filter removal was hidden or forgotten")
	}
	b.failRemove = false
	m.sweepReactive()
	if b.Count() != 0 || m.reactive.Active() != 0 || m.Health().EnforcementDegraded {
		t.Fatal("exempt filter was handed to raw feed or failed to retry")
	}
	// Adding a host resolver/server exemption uses the same immediate cleanup.
	cfg.Allowlist.Networks = nil
	m.Reload(cfg, SystemNets{})
	m.reactive.Block([]netip.Addr{a})
	m.Reload(cfg, SystemNets{Resolvers: []netip.Addr{a}})
	if b.Count() != 0 || m.reactive.Active() != 0 {
		t.Fatal("resolver remained reactively blocked")
	}
}

func TestReactiveExpiryAndDisableFailuresRemainTracked(t *testing.T) {
	b := &retryBlocker{fakeBlocker: newFake()}
	m, cfg, _ := reactiveManager(t, b)
	clock := time.Unix(1000, 0)
	m.reactive.now = func() time.Time { return clock }
	a := netip.MustParseAddr("192.0.2.6")
	m.onDNSEvent(DNSEvent{Name: "evil.example", IPs: []netip.Addr{a}})
	clock = clock.Add(time.Hour)
	b.failRemove = true
	m.sweepReactive()
	if m.reactive.Active() != 1 || b.Count() != 1 || !m.Health().EnforcementDegraded {
		t.Fatal("failed expiry forgotten")
	}
	b.failRemove = false
	m.sweepReactive()
	if m.reactive.Active() != 0 || b.Count() != 0 || m.Health().EnforcementDegraded {
		t.Fatal("expiry retry did not recover")
	}
	m.onDNSEvent(DNSEvent{Name: "evil.example", IPs: []netip.Addr{a}})
	b.failReset = true
	cfg.Blocklist.Enabled = false
	m.Reload(cfg, SystemNets{})
	if m.reactive.Active() != 1 || b.Count() != 1 || !m.Health().EnforcementDegraded {
		t.Fatal("failed disable reset forgotten")
	}
	b.failReset = false
	m.sweepReactive()
	if m.reactive.Active() != 0 || b.Count() != 0 || m.Health().EnforcementDegraded {
		t.Fatal("disable cleanup did not retry")
	}
}

func TestReactiveSharedOwnershipSurvivesFeedRemoval(t *testing.T) {
	b := newFake()
	m, _, _ := reactiveManager(t, b)
	clock := time.Unix(1000, 0)
	m.reactive.now = func() time.Time { return clock }
	a := netip.MustParseAddr("192.0.2.7")
	m.onDNSEvent(DNSEvent{Name: "evil.example", IPs: []netip.Addr{a}})
	ip, _ := ParseIndicator(a.String(), "ip", 1)
	m.applyIndicators([]Indicator{ip})
	m.applyIndicators(nil)
	if b.Count() != 1 || m.enf.OwnsIP(a) || !m.reactive.Owns(a) {
		t.Fatal("feed removal deleted a live reactive lease")
	}
	clock = clock.Add(time.Hour)
	m.sweepReactive()
	if b.Count() != 0 || m.reactive.Active() != 0 {
		t.Fatal("released feed filter did not expire")
	}
	// A feed that actually owns the exact filter keeps it at expiry.
	m.reactive.Block([]netip.Addr{a})
	m.applyIndicators([]Indicator{ip})
	clock = clock.Add(time.Hour)
	m.sweepReactive()
	if b.Count() != 1 || m.reactive.Active() != 0 || !m.enf.OwnsIP(a) {
		t.Fatal("actual feed ownership did not retain its filter")
	}
}

func TestReactiveCapacityAndMixedCoverageAreHonest(t *testing.T) {
	b := newFake()
	m, _, sp := reactiveManager(t, b)
	m.reactive.limit = 1
	a, c := netip.MustParseAddr("192.0.2.8"), netip.MustParseAddr("192.0.2.9")
	m.onDNSEvent(DNSEvent{Name: "evil.example", IPs: []netip.Addr{a, c}})
	if b.Count() != 1 || m.reactive.Active() != 1 || !errors.Is(m.reactive.HealthError(), ErrReactiveBudget) ||
		!strings.Contains(sp.lines[len(sp.lines)-1], `"action":"detected"`) {
		t.Fatal("capacity failure claimed complete blocking")
	}
	// Already-enforced feed IPs and a successful new reactive IP jointly cover
	// all answers; unlike raw membership, actual ownership earns this outcome.
	ip, _ := ParseIndicator(a.String(), "ip", 1)
	m.applyIndicators([]Indicator{ip})
	m.reactive.limit = 2
	m.onDNSEvent(DNSEvent{Name: "evil.example", IPs: []netip.Addr{a, c}})
	if !strings.Contains(sp.lines[len(sp.lines)-1], `"action":"blocked"`) {
		t.Fatal("complete shared coverage was not recognized")
	}
}

func TestInboundOnlyFiltersDoNotClaimOutboundDNSBlocking(t *testing.T) {
	b := newFake()
	m, _, sp := reactiveManager(t, b)
	m.enf.dir, m.reactive.dir = "in", "in"
	// A structural direction edit has not rebuilt the installed backend yet.
	m.cfg.Blocklist.Direction = "out"
	m.onDNSEvent(DNSEvent{Name: "evil.example", IPs: []netip.Addr{netip.MustParseAddr("192.0.2.11")}})
	if b.Count() != 0 || !strings.Contains(sp.lines[0], `"action":"detected"`) {
		t.Fatal("inbound-only filters claimed outbound DNS protection")
	}
}

func TestFailedRemovalCannotClaimCompleteEnforcerOwnership(t *testing.T) {
	b := &retryBlocker{fakeBlocker: newFake()}
	e := NewEnforcer(b, "both")
	a := netip.MustParseAddr("192.0.2.12")
	if err := e.Apply([]netip.Addr{a}); err != nil {
		t.Fatal(err)
	}
	b.failRemove = true
	if err := e.Apply(nil); err == nil || e.OwnsIP(a) || e.BlocksIP(a) {
		t.Fatal("possibly partial removal still claimed complete protection")
	}
	b.failRemove = false
	if err := e.Apply([]netip.Addr{a}); err != nil || !e.OwnsIP(a) {
		t.Fatal("desired filter was not revalidated after a partial removal")
	}
}

func TestReactiveManagerLockOrderUnderConcurrentReconciliation(t *testing.T) {
	b := newFake()
	m, _, _ := reactiveManager(t, b)
	a := netip.MustParseAddr("192.0.2.10")
	var wg sync.WaitGroup
	for _, work := range []func(){
		func() { m.onDNSEvent(DNSEvent{Name: "evil.example", IPs: []netip.Addr{a}}) },
		func() { m.sweepReactive() },
		func() { m.applyIndicators(nil) },
		func() { _ = m.Health() },
	} {
		wg.Add(1)
		go func(f func()) {
			defer wg.Done()
			for i := 0; i < 30; i++ {
				f()
			}
		}(work)
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("ownership lock order deadlocked")
	}
}
