//go:build linux

package netblock

import (
	"context"
	"errors"
	"net/netip"
	"path/filepath"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
)

type safetyRecordingBlocker struct {
	*fakeBlocker
	policy  SafetyPolicy
	updates int
}

func (b *safetyRecordingBlocker) SetSafetyPolicy(p SafetyPolicy) error {
	b.updates++
	b.policy = p
	if err := managementValid(p.Management); err != nil {
		_ = b.Reset()
		return err
	}
	return nil
}

func systemNetManager(t *testing.T) (*Manager, *safetyRecordingBlocker, []netip.Addr) {
	t.Helper()
	c, err := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	cfg := config.Default()
	cfg.Server = "original.openedr.test"
	cfg.Blocklist.Direction = "out"
	peers := []netip.Addr{netip.MustParseAddr("192.0.2.1"), netip.MustParseAddr("198.51.100.1"), netip.MustParseAddr("203.0.113.1")}
	b := &safetyRecordingBlocker{fakeBlocker: newFake()}
	m := NewManager(Deps{Cfg: cfg, Cache: c, Spool: &memSpool{}, Sys: SystemNets{ServerHosts: peers[:1]}, NewBlocker: func(string) (Blocker, error) { return b, nil }})
	var indicators []Indicator
	for _, peer := range peers {
		ind, _ := ParseIndicator(peer.String(), "ip", 1)
		indicators = append(indicators, ind)
	}
	m.store.Swap(indicators)
	m.applyIndicators(indicators)
	return m, b, peers
}

func TestLinuxSystemNetRefreshUpdatesSafetyWithoutReloadingFeed(t *testing.T) {
	m, b, peers := systemNetManager(t)
	if got := b.keys(); !reflect.DeepEqual(got, []string{"198.51.100.1", "203.0.113.1"}) {
		t.Fatal(got)
	}
	var cancellations atomic.Int32
	m.feed.activeCancel = func() { cancellations.Add(1) }
	syncTime := time.Now()
	m.feed.lastFullSync = syncTime
	resolve := func(context.Context, config.EDRConfig) SystemNets { return SystemNets{ServerHosts: peers[1:2]} }
	applied, err := m.RefreshSystemNets(context.Background(), resolve)
	if err != nil || !applied {
		t.Fatal(applied, err)
	}
	if got := b.keys(); !reflect.DeepEqual(got, []string{"192.0.2.1", "203.0.113.1"}) {
		t.Fatal("new management address still blocked", got)
	}
	before := b.updates
	applied, err = m.RefreshSystemNets(context.Background(), resolve)
	if err != nil || applied || b.updates != before {
		t.Fatal("unchanged healthy DNS response churned policy", applied, err, b.updates, before)
	}
	applied, err = m.RefreshSystemNets(context.Background(), func(context.Context, config.EDRConfig) SystemNets { return SystemNets{} })
	if !applied || !errors.Is(err, ErrNetworkSafety) || b.Count() != 0 || m.EnforcementHealth() == nil {
		t.Fatal("unavailable management retained enforcement or healthy status", applied, err, b.Count())
	}
	applied, err = m.RefreshSystemNets(context.Background(), resolve)
	if err != nil || !applied || m.EnforcementHealth() != nil || !reflect.DeepEqual(b.keys(), []string{"192.0.2.1", "203.0.113.1"}) {
		t.Fatal("recovered lookup did not restore indicator policy", applied, err, b.keys())
	}
	if cancellations.Load() != 0 || len(m.feed.reloadCh) != 0 || !m.feed.lastFullSync.Equal(syncTime) {
		t.Fatal("safety refresh restarted/cancelled a feed transaction")
	}
}

func TestLinuxSystemNetRefreshDiscardsLookupFromReplacedConfiguration(t *testing.T) {
	m, b, peers := systemNetManager(t)
	m.mu.Lock()
	next := m.cfg
	m.mu.Unlock()
	next.Server = "replacement.openedr.test"
	entered, release := make(chan struct{}), make(chan struct{})
	result := make(chan error, 1)
	go func() {
		applied, err := m.RefreshSystemNets(context.Background(), func(_ context.Context, cfg config.EDRConfig) SystemNets {
			if cfg.Server != "original.openedr.test" {
				t.Errorf("wrong lookup configuration: %s", cfg.Server)
			}
			close(entered)
			<-release
			return SystemNets{ServerHosts: peers[1:2]}
		})
		if applied && err == nil {
			err = errors.New("stale lookup applied")
		}
		result <- err
	}()
	<-entered
	reloaded := make(chan struct{})
	go func() { m.Reload(next, SystemNets{ServerHosts: peers[:1]}); close(reloaded) }()
	select {
	case <-reloaded:
	case <-time.After(time.Second):
		close(release)
		<-result
		<-reloaded
		t.Fatal("DNS lookup held the configuration policy lock")
	}
	before := b.updates
	close(release)
	if err := <-result; err != nil || b.updates != before || !reflect.DeepEqual(b.policy.Management, peers[:1]) {
		t.Fatal("old lookup overrode new configuration", err, b.policy.Management)
	}
}

func TestLinuxSystemNetRefreshCancellationDoesNotMutatePolicy(t *testing.T) {
	m, b, peers := systemNetManager(t)
	ctx, cancel := context.WithCancel(context.Background())
	entered := make(chan struct{})
	result := make(chan error, 1)
	before := b.updates
	go func() {
		_, err := m.RefreshSystemNets(ctx, func(ctx context.Context, _ config.EDRConfig) SystemNets {
			close(entered)
			<-ctx.Done()
			return SystemNets{ServerHosts: peers[1:2]}
		})
		result <- err
	}()
	<-entered
	cancel()
	if err := <-result; !errors.Is(err, context.Canceled) || b.updates != before {
		t.Fatal("cancelled lookup changed policy", err, b.updates, before)
	}
}

func TestLinuxNetworkConstructorStaysUnreadyUntilCachedPolicyReconciles(t *testing.T) {
	c, err := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	cfg := config.Default()
	cfg.LinuxProtection.NetworkEnforcement = true
	b := &safetyRecordingBlocker{fakeBlocker: newFake()}
	m := NewManager(Deps{Cfg: cfg, Cache: c, Spool: &memSpool{}, Sys: SystemNets{ServerHosts: []netip.Addr{netip.MustParseAddr("192.0.2.1")}}, NewBlocker: func(string) (Blocker, error) { return b, nil }})
	if m.EnforcementHealth() == nil {
		t.Fatal("constructor alone reported ready before cached policy reconciliation")
	}
	applied, err := m.RefreshSystemNets(context.Background(), func(context.Context, config.EDRConfig) SystemNets {
		return SystemNets{ServerHosts: []netip.Addr{netip.MustParseAddr("192.0.2.1")}}
	})
	if !applied || err != nil || m.EnforcementHealth() != nil {
		t.Fatal("successfully reconciled empty cache remained unready", applied, err)
	}
}

func TestLinuxSystemNetRefreshDiscardsLookupAfterIdenticalConfigurationReload(t *testing.T) {
	m, b, peers := systemNetManager(t)
	m.mu.Lock()
	same := m.cfg
	m.mu.Unlock()
	entered, release := make(chan struct{}), make(chan struct{})
	result := make(chan bool, 1)
	go func() {
		applied, err := m.RefreshSystemNets(context.Background(), func(context.Context, config.EDRConfig) SystemNets {
			close(entered)
			<-release
			return SystemNets{ServerHosts: peers[1:2]}
		})
		result <- applied || err != nil
	}()
	<-entered
	m.Reload(same, SystemNets{ServerHosts: peers[:1]})
	before := b.updates
	close(release)
	if changed := <-result; changed || b.updates != before || !reflect.DeepEqual(b.policy.Management, peers[:1]) {
		t.Fatal("in-flight lookup overwrote an authoritative reload of identical configuration", changed, b.updates, before, b.policy.Management)
	}
}
