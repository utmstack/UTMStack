package usecase

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
)

type fakeAlerts struct {
	buckets []connectors.FloodBucket
	err     error
	window  time.Duration
}

func (f *fakeAlerts) FloodBuckets(_ context.Context, window time.Duration) ([]connectors.FloodBucket, error) {
	f.window = window

	return f.buckets, f.err
}

type fakeRules struct {
	mu          sync.Mutex
	byTenant    map[string][]connectors.FloodRule
	findErr     error
	deactivate  map[string]bool // relPath -> what Deactivate reports as changed
	deactErr    map[string]error
	deactivated []string
}

func (f *fakeRules) FindByName(_ context.Context, tenantID, _ string) ([]connectors.FloodRule, error) {
	return f.byTenant[tenantID], f.findErr
}

func (f *fakeRules) Deactivate(_ context.Context, tenantID, relPath string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.deactErr[relPath]; err != nil {
		return false, err
	}

	f.deactivated = append(f.deactivated, tenantID+":"+relPath)

	return f.deactivate[relPath], nil
}

type fakeNotifier struct {
	sent []string
	err  error
}

func (f *fakeNotifier) Warn(_ context.Context, tenantID, message string) error {
	f.sent = append(f.sent, tenantID+"|"+message)

	return f.err
}

func cfg(enabled bool, threshold int64) func() FloodConfig {
	c := DefaultFloodConfig()
	c.Enabled = enabled
	c.Threshold = threshold

	return func() FloodConfig { return c }
}

func guard(a *fakeAlerts, r *fakeRules, n *fakeNotifier, c func() FloodConfig) *FloodGuard {
	return NewFloodGuard(a, r, n, c)
}

func rule(relPath, name string, active bool) connectors.FloodRule {
	return connectors.FloodRule{RelPath: relPath, Name: name, Active: active}
}

func TestOnlyBucketsOverTheThresholdDisableTheirRule(t *testing.T) {
	a := &fakeAlerts{buckets: []connectors.FloodBucket{
		{TenantID: "t1", RuleName: "Noisy", DataSource: "fw", Count: 51},
		{TenantID: "t1", RuleName: "Quiet", DataSource: "fw", Count: 50},
	}}
	r := &fakeRules{
		byTenant:   map[string][]connectors.FloodRule{"t1": {rule("noisy.yaml", "Noisy", true), rule("quiet.yaml", "Quiet", true)}},
		deactivate: map[string]bool{"noisy.yaml": true, "quiet.yaml": true},
	}
	n := &fakeNotifier{}

	guard(a, r, n, cfg(true, 50)).evaluateOnce(context.Background())

	if len(r.deactivated) != 1 || r.deactivated[0] != "t1:noisy.yaml" {
		t.Errorf("deactivated %v, want only the rule above the threshold", r.deactivated)
	}

	if len(n.sent) != 1 || !strings.Contains(n.sent[0], "'Noisy'") || !strings.Contains(n.sent[0], "51 open") || !strings.Contains(n.sent[0], "'fw'") {
		t.Errorf("notification: %v", n.sent)
	}
}

func TestTheWindowComesFromTheConfiguration(t *testing.T) {
	a := &fakeAlerts{}
	c := DefaultFloodConfig()
	c.WindowHours = 6

	guard(a, &fakeRules{}, &fakeNotifier{}, func() FloodConfig { return c }).evaluateOnce(context.Background())

	if a.window != 6*time.Hour {
		t.Errorf("window %v, want 6h", a.window)
	}
}

func TestDisabledGuardDoesNotEvenLookAtTheAlerts(t *testing.T) {
	a := &fakeAlerts{buckets: []connectors.FloodBucket{{TenantID: "t1", RuleName: "Noisy", Count: 1000}}}

	guard(a, &fakeRules{}, &fakeNotifier{}, cfg(false, 1)).evaluateOnce(context.Background())

	if a.window != 0 {
		t.Error("a disabled guard must not query")
	}
}

func TestOnlyTheCallThatChangedTheRuleNotifies(t *testing.T) {
	a := &fakeAlerts{buckets: []connectors.FloodBucket{{TenantID: "t1", RuleName: "Noisy", DataSource: "fw", Count: 99}}}
	r := &fakeRules{
		byTenant:   map[string][]connectors.FloodRule{"t1": {rule("noisy.yaml", "Noisy", true)}},
		deactivate: map[string]bool{"noisy.yaml": false}, // another replica got there first
	}
	n := &fakeNotifier{}

	guard(a, r, n, cfg(true, 50)).evaluateOnce(context.Background())

	if len(n.sent) != 0 {
		t.Errorf("a call that did not change the rule must not notify: %v", n.sent)
	}
}

func TestEachTenantIsHandledOnItsOwn(t *testing.T) {
	a := &fakeAlerts{buckets: []connectors.FloodBucket{
		{TenantID: "t1", RuleName: "Noisy", DataSource: "fw", Count: 60},
		{TenantID: "t2", RuleName: "Noisy", DataSource: "fw", Count: 10},
	}}
	r := &fakeRules{
		byTenant: map[string][]connectors.FloodRule{
			"t1": {rule("t1/noisy.yaml", "Noisy", true)},
			"t2": {rule("t2/noisy.yaml", "Noisy", true)},
		},
		deactivate: map[string]bool{"t1/noisy.yaml": true, "t2/noisy.yaml": true},
	}
	n := &fakeNotifier{}

	guard(a, r, n, cfg(true, 50)).evaluateOnce(context.Background())

	if len(r.deactivated) != 1 || r.deactivated[0] != "t1:t1/noisy.yaml" {
		t.Errorf("only the tenant that flooded loses the rule: %v", r.deactivated)
	}

	if len(n.sent) != 1 || !strings.HasPrefix(n.sent[0], "t1|") {
		t.Errorf("only the tenant that flooded is told: %v", n.sent)
	}
}

func TestABucketWithoutATenantIsDropped(t *testing.T) {
	a := &fakeAlerts{buckets: []connectors.FloodBucket{{TenantID: "", RuleName: "Noisy", Count: 500}}}
	r := &fakeRules{byTenant: map[string][]connectors.FloodRule{"": {rule("noisy.yaml", "Noisy", true)}}, deactivate: map[string]bool{"noisy.yaml": true}}

	guard(a, r, &fakeNotifier{}, cfg(true, 50)).evaluateOnce(context.Background())

	if len(r.deactivated) != 0 {
		t.Errorf("an alert with no tenant must not be attributed to any: %v", r.deactivated)
	}
}

func TestOnlyExactNameMatchesAreDisabledAndInactiveOnesAreLeftAlone(t *testing.T) {
	a := &fakeAlerts{buckets: []connectors.FloodBucket{{TenantID: "t1", RuleName: "Port Scan", DataSource: "fw", Count: 80}}}
	r := &fakeRules{
		byTenant: map[string][]connectors.FloodRule{"t1": {
			rule("a.yaml", "port scan", true),          // same name, other case
			rule("b.yaml", "Port Scan Extended", true), // partial match only
			rule("c.yaml", "Port Scan", false),         // already off
		}},
		deactivate: map[string]bool{"a.yaml": true, "b.yaml": true, "c.yaml": true},
	}

	guard(a, r, &fakeNotifier{}, cfg(true, 50)).evaluateOnce(context.Background())

	if len(r.deactivated) != 1 || r.deactivated[0] != "t1:a.yaml" {
		t.Errorf("deactivated %v, want only the exact active match", r.deactivated)
	}
}

func TestAFailureOnOneBucketDoesNotStopTheNext(t *testing.T) {
	a := &fakeAlerts{buckets: []connectors.FloodBucket{
		{TenantID: "t1", RuleName: "Broken", DataSource: "fw", Count: 90},
		{TenantID: "t2", RuleName: "Fine", DataSource: "fw", Count: 90},
	}}
	r := &fakeRules{
		byTenant: map[string][]connectors.FloodRule{
			"t1": {rule("broken.yaml", "Broken", true)},
			"t2": {rule("fine.yaml", "Fine", true)},
		},
		deactivate: map[string]bool{"fine.yaml": true},
		deactErr:   map[string]error{"broken.yaml": errors.New("boom")},
	}
	n := &fakeNotifier{err: errors.New("bell down")}

	guard(a, r, n, cfg(true, 50)).evaluateOnce(context.Background())

	if len(r.deactivated) != 1 || r.deactivated[0] != "t2:fine.yaml" {
		t.Errorf("the second bucket must still be handled: %v", r.deactivated)
	}

	if len(n.sent) != 1 {
		t.Errorf("a failed notification is reported, not fatal: %v", n.sent)
	}
}

func TestAFailedCountOrAPanicLeavesTheGuardRunning(t *testing.T) {
	failing := &fakeAlerts{err: errors.New("clickhouse down")}
	guard(failing, &fakeRules{}, &fakeNotifier{}, cfg(true, 50)).evaluateOnce(context.Background())

	g := NewFloodGuard(&fakeAlerts{}, &fakeRules{}, &fakeNotifier{}, func() FloodConfig { panic("config exploded") })
	g.evaluateOnce(context.Background()) // must not panic out
}

func TestStartRunsImmediatelyAndStopsWithTheContext(t *testing.T) {
	a := &fakeAlerts{}
	c := DefaultFloodConfig()
	c.Interval = 20 * time.Millisecond

	g := guard(a, &fakeRules{}, &fakeNotifier{}, func() FloodConfig { return c })

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() { g.Start(ctx); close(done) }()

	time.Sleep(80 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Start must return when the context is done")
	}

	if a.window == 0 {
		t.Error("the guard must have checked at least once")
	}
}

func TestNilGuardStartIsHarmless(t *testing.T) {
	var g *FloodGuard

	g.Start(context.Background())
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, floodConfigFile), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	return dir
}

func TestConfigWithoutAFileUsesTheDefaults(t *testing.T) {
	got, err := LoadFloodConfig(t.TempDir())
	if err != nil || got != DefaultFloodConfig() {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestConfigReadsWhatTheFileSetsAndKeepsTheRest(t *testing.T) {
	dir := writeConfig(t, "plugins:\n  rule-flood-guard:\n    enabled: false\n    threshold: 10\n    windowHours: 2\n")

	got, err := LoadFloodConfig(dir)
	if err != nil {
		t.Fatal(err)
	}

	if got.Enabled || got.Threshold != 10 || got.WindowHours != 2 || got.Interval != 300*time.Second {
		t.Errorf("unexpected: %+v", got)
	}
}

func TestConfigIgnoresValuesThatAreNotPositive(t *testing.T) {
	dir := writeConfig(t, "plugins:\n  rule-flood-guard:\n    threshold: 0\n    windowHours: -3\n    intervalSeconds: 0\n")

	got, err := LoadFloodConfig(dir)
	if err != nil || got != DefaultFloodConfig() {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestConfigWithABrokenFileFallsBackToDefaultsAndSaysWhy(t *testing.T) {
	got, err := LoadFloodConfig(writeConfig(t, "plugins: [not a map"))
	if err == nil {
		t.Error("a file that cannot be read must be reported")
	}

	if got != DefaultFloodConfig() {
		t.Errorf("must fall back to the defaults: %+v", got)
	}
}
