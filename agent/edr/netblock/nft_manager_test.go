package netblock

import (
	"context"
	"errors"
	"net/netip"
	"path/filepath"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
)

type lifecycleBlocker struct {
	*fakeBlocker
	stopped   chan struct{}
	closed    chan bool
	healthErr error
}

func (b *lifecycleBlocker) Run(ctx context.Context) { <-ctx.Done(); close(b.stopped) }
func (b *lifecycleBlocker) Close() error {
	select {
	case <-b.stopped:
		b.closed <- true
	default:
		b.closed <- false
	}
	return nil
}
func (b *lifecycleBlocker) Health() error { return b.healthErr }

func TestManagerLeaseLifecycleClosesOnlyAfterRefresherStopped(t *testing.T) {
	c, err := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	b := &lifecycleBlocker{fakeBlocker: newFake(), stopped: make(chan struct{}), closed: make(chan bool, 1)}
	cfg := config.Default()
	cfg.Server = "" // no network feed, even if Run is driven
	m := NewManager(Deps{Cfg: cfg, Cache: c, Spool: &memSpool{}, NewBlocker: func(string) (Blocker, error) { return b, nil }})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan struct{})
	go func() { m.Run(ctx); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("manager lifecycle did not stop")
	}
	if !<-b.closed {
		t.Fatal("backend closed before lease refresher stopped")
	}
}

func TestEnforcementHealthDoesNotConsumeEventsAndDegradedDNSCannotClaimBlocked(t *testing.T) {
	c, err := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	b := &lifecycleBlocker{fakeBlocker: newFake()}
	m := NewManager(Deps{Cfg: config.Default(), Cache: c, Spool: &memSpool{}, NewBlocker: func(string) (Blocker, error) { return b, nil }})
	m.recentBlocks.Store(7)
	if err = m.EnforcementHealth(); err != nil || m.recentBlocks.Load() != 7 {
		t.Fatalf("side effect from health: %v %d", err, m.recentBlocks.Load())
	}
	if got := m.Health().RecentBlocks; got != 7 {
		t.Fatal(got)
	}
	ip := netip.MustParseAddr("192.0.2.9")
	ind, _ := ParseIndicator("test.invalid", "domain", 1)
	m.store.Swap([]Indicator{ind})
	if err = m.enf.Apply([]netip.Addr{ip}); err != nil {
		t.Fatal(err)
	}
	b.healthErr = errors.New("lease refresh failed")
	if m.EnforcementHealth() == nil {
		t.Fatal("backend lease failure hidden")
	}
	m.onDNSEvent(DNSEvent{Name: "test.invalid", IPs: []netip.Addr{ip}})
	if m.recentBlocks.Load() != 0 {
		t.Fatal("unverified existing filter earned a blocked event")
	}
}

func TestNFTBlockerDoesNotPublishFailedDesiredState(t *testing.T) {
	c, f := newFakeNFT(t)
	b := newNFTBlocker(c, nil)
	a := netip.MustParseAddr("192.0.2.9")
	z := netip.MustParseAddr("192.0.2.10")
	if err := b.AddIP(a, "both"); !errors.Is(err, ErrNetworkSafety) || len(f.transactions) != 0 {
		t.Fatalf("unguarded enforcement %v", err)
	}
	if err := b.SetSafetyPolicy(testSafety()); err != nil {
		t.Fatal(err)
	}
	p, _ := blockPlan(c.owner, testSafety(), map[netip.Addr]string{a: "both"}, nil, f.now)
	f.register(p)
	if err := b.AddIP(a, "both"); err != nil {
		t.Fatal(err)
	}
	q, _ := blockPlan(c.owner, testSafety(), map[netip.Addr]string{a: "both", z: "out"}, nil, f.now)
	f.register(q)
	f.failAfter = 1
	if err := b.AddIP(z, "out"); err == nil {
		t.Fatal("post-commit error hidden")
	}
	if len(b.ips) != 1 || b.ips[a] != "both" || b.Health() == nil {
		t.Fatal("failed state published or error forgotten")
	}
	if err := b.AddIP(z, "out"); err != nil || len(b.ips) != 2 || b.Health() != nil {
		t.Fatalf("retry failed: %v", err)
	}
	if err := b.SetSafetyPolicy(SafetyPolicy{}); !errors.Is(err, ErrNetworkSafety) || f.current != nil || b.Count() != 0 {
		t.Fatalf("lost management did not fail open: %v", err)
	}
}

func TestNFTInitialSafetyRemovesStaleOwnedTableBeforeReportingEmptyHealthy(t *testing.T) {
	c, f := newFakeNFT(t)
	p, _ := blockPlan(c.owner, testSafety(), map[netip.Addr]string{netip.MustParseAddr("192.0.2.8"): "both"}, nil, f.now)
	f.register(p)
	f.current = p
	b := newNFTBlocker(c, nil)
	if err := b.SetSafetyPolicy(testSafety()); err != nil || f.current != nil || b.Count() != 0 || b.Health() != nil {
		t.Fatalf("stale owned enforcement survived initial convergence: %v", err)
	}
}

func TestManagerMinuteSweepRetriesOnlyLatestCommittedDesiredAfterOSFailure(t *testing.T) {
	c, err := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	b := &batchFakeBlocker{fakeBlocker: newFake()}
	m := NewManager(Deps{Cfg: config.Default(), Cache: c, Spool: &memSpool{}, NewBlocker: func(string) (Blocker, error) { return b, nil }})
	a, _ := ParseIndicator("192.0.2.8", "ip", 1)
	z, _ := ParseIndicator("192.0.2.9", "ip", 1)
	m.store.Swap([]Indicator{a})
	m.applyIndicators([]Indicator{a})
	b.fail = true
	m.store.Swap([]Indicator{z})
	m.applyIndicators([]Indicator{z})
	if m.EnforcementHealth() == nil {
		t.Fatal("failed OS apply did not degrade health")
	}
	m.lastSync.Store(100)
	m.sweepReactive()
	if m.EnforcementHealth() == nil {
		t.Fatal("repeated OS failure became healthy")
	}
	b.fail = false
	m.sweepReactive()
	if m.EnforcementHealth() != nil || !m.enf.OwnsIP(netip.MustParseAddr(z.Value)) || m.enf.OwnsIP(netip.MustParseAddr(a.Value)) {
		t.Fatal("minute retry did not converge the latest committed desired set")
	}
	if m.lastSync.Load() != 100 {
		t.Fatal("OS-only retry falsely advanced feed freshness")
	}
	before := b.calls
	m.feed.recordHealth(errors.New("malformed feed rejected"))
	m.sweepReactive()
	if b.calls != before || !m.Health().FeedStale {
		t.Fatal("malformed feed caused an OS desired-policy replacement or hid stale feed")
	}
}
