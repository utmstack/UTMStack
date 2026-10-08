//go:build windows || linux

package ransomware

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
)

// newFuzzyGuard builds a kill-mode guard (canary-free) with an injectable
// EntropyReader and, optionally, a movable clock so several signals can be
// spaced within one decay half-life.
func newFuzzyGuard(t *testing.T, reader FileReader, movableClock bool) (*Guard, *fakeResp, *fakeIncidents, *proctable.Table) {
	t.Helper()
	tab := proctable.New()
	resp := &fakeResp{}
	inc := &fakeIncidents{}
	var clock int64
	cfg := config.Default()
	cfg.Ransomware.Enabled = config.BoolPtr(true)
	cfg.Ransomware.ResponseMode = "kill"
	cfg.Ransomware.SuspendThreshold = 50
	cfg.Ransomware.KillThreshold = 100
	now := func() time.Time { return time.Unix(1000, 0) }
	if movableClock {
		now = func() time.Time { return time.Unix(clock, 0) }
	}
	g := NewGuard(GuardDeps{
		Cfg: cfg, Table: tab, Resp: resp, Spool: &fakeAppender{}, Quar: &fakeQuar{}, Incidents: inc,
		Canaries:      &fakeCanaries{paths: map[string]bool{}},
		Hash:          func(string) (string, error) { return "deadbeef", nil },
		Now:           now,
		NewID:         func() string { return "inc-1" },
		EntropyReader: reader,
	})
	return g, resp, inc, tab
}

// entropyReader serves 4096 bytes of uniform 256-symbol content (8 bits/byte)
// so the entropy sensor fires deterministically for any path.
func entropyReader(path string, offset int64, size int) ([]byte, error) {
	buf := make([]byte, size)
	for i := range buf {
		buf[i] = byte(i)
	}
	return buf, nil
}

// TestGuardFuzzy_SingleSignalNeverEscalates verifies the core X1 invariant:
// each fuzzy sensor, even when it fires, stays below the suspend threshold on
// its own. Only the kind-diversity bonus lets several agree into an escalation.
func TestGuardFuzzy_SingleSignalNeverEscalates(t *testing.T) {
	// 1) Ransom note alone (weight 30).
	g, resp, inc, tab := newFuzzyGuard(t, nil, false)
	tab.Add(proctable.Proc{PID: 300, PPID: 4, Image: `/bin/enc`})
	g.OnFileEvent(FileEvent{PID: 300, Path: `/data/README_RECOVER.txt`, Op: OpWrite})
	if len(resp.killed) != 0 || len(resp.suspended) != 0 || len(inc.recs) != 0 {
		t.Fatalf("ransom note alone must not escalate: killed=%v suspended=%v recs=%d",
			resp.killed, resp.suspended, len(inc.recs))
	}

	// 2) Churn alone (weight 40): 12 distinct unusual-extension writes in one
	// tick (defaults: ChurnMinRenames=10, ChurnMinDistinct=5).
	g, resp, inc, tab = newFuzzyGuard(t, nil, false)
	tab.Add(proctable.Proc{PID: 300, PPID: 4, Image: `/bin/enc`})
	for i := 0; i < 12; i++ {
		g.OnFileEvent(FileEvent{PID: 300, Path: "/data/c" + strings.Repeat("x", i) + ".locked", Op: OpWrite})
	}
	if len(resp.killed) != 0 || len(resp.suspended) != 0 || len(inc.recs) != 0 {
		t.Fatalf("churn alone must not escalate: killed=%v suspended=%v recs=%d",
			resp.killed, resp.suspended, len(inc.recs))
	}

	// 3) Mass rate alone (weight 35): 26 writes in one tick (default
	// MassRateMinWrites=25). .log is a common extension so churn stays blind.
	g, resp, inc, tab = newFuzzyGuard(t, nil, false)
	tab.Add(proctable.Proc{PID: 300, PPID: 4, Image: `/bin/enc`})
	for i := 0; i < 26; i++ {
		g.OnFileEvent(FileEvent{PID: 300, Path: "/data/m" + strings.Repeat("y", i) + ".log", Op: OpWrite})
	}
	if len(resp.killed) != 0 || len(resp.suspended) != 0 || len(inc.recs) != 0 {
		t.Fatalf("mass rate alone must not escalate: killed=%v suspended=%v recs=%d",
			resp.killed, resp.suspended, len(inc.recs))
	}

	// 4) Entropy alone (weight 40): one write to a high-entropy file, with an
	// injected reader. The sensor stats the real path first, so create it on
	// disk (>= sample size so it samples multiple offsets); the injected reader
	// serves the synthetic high-entropy content regardless of on-disk bytes.
	g, resp, inc, tab = newFuzzyGuard(t, entropyReader, false)
	tab.Add(proctable.Proc{PID: 300, PPID: 4, Image: `/bin/enc`})
	entPath := t.TempDir() + "/e0.bin"
	if err := os.WriteFile(entPath, make([]byte, 8192), 0o600); err != nil {
		t.Fatalf("write entropy fixture: %v", err)
	}
	g.OnFileEvent(FileEvent{PID: 300, Path: entPath, Op: OpWrite})
	if len(resp.killed) != 0 || len(resp.suspended) != 0 || len(inc.recs) != 0 {
		t.Fatalf("entropy alone must not escalate: killed=%v suspended=%v recs=%d",
			resp.killed, resp.suspended, len(inc.recs))
	}
}

// TestGuardFuzzy_ThreeSignalsCrossKill verifies the X1 design goal: three
// distinct fuzzy kinds firing within one decay half-life cross the kill
// threshold via the kind-diversity bonus (1.25 x 1.5), even though each
// signal alone stays under the suspend threshold.
func TestGuardFuzzy_ThreeSignalsCrossKill(t *testing.T) {
	tab := proctable.New()
	resp := &fakeResp{}
	inc := &fakeIncidents{}
	var clock int64
	cfg := config.Default()
	cfg.Ransomware.Enabled = config.BoolPtr(true)
	cfg.Ransomware.ResponseMode = "kill"
	cfg.Ransomware.SuspendThreshold = 50
	cfg.Ransomware.KillThreshold = 100
	g := NewGuard(GuardDeps{
		Cfg: cfg, Table: tab, Resp: resp, Spool: &fakeAppender{}, Quar: &fakeQuar{}, Incidents: inc,
		Canaries: &fakeCanaries{paths: map[string]bool{}},
		Hash:     func(string) (string, error) { return "deadbeef", nil },
		Now:      func() time.Time { return time.Unix(clock, 0) },
		NewID:    func() string { return "inc-1" },
	})
	tab.Add(proctable.Proc{PID: 300, PPID: 4, Image: `/bin/enc`})

	// 1) Ransom note at t=1000 (score 30).
	clock = 1000
	g.OnFileEvent(FileEvent{PID: 300, Path: `/data/README_RECOVER.txt`, Op: OpWrite})

	// 2) Churn at t=1004 (12 unusual writes). Score decays ~23 + 40 = ~63;
	// with two kinds eff ~ 78 -> EscSuspend, no action in kill mode.
	clock = 1004
	for i := 0; i < 12; i++ {
		g.OnFileEvent(FileEvent{PID: 300, Path: "/data/b" + strings.Repeat("x", i) + ".locked", Op: OpWrite})
	}
	if len(resp.killed) != 0 {
		t.Fatalf("two fuzzy signals must not kill yet: killed=%v", resp.killed)
	}

	// 3) Mass rate at t=1009 (26 common-ext writes, invisible to churn).
	// Score decays to ~44 + 35 = ~79; with three kinds eff ~ 118 -> EscKill.
	clock = 1009
	for i := 0; i < 26; i++ {
		g.OnFileEvent(FileEvent{PID: 300, Path: "/data/m" + strings.Repeat("y", i) + ".log", Op: OpWrite})
	}
	if len(resp.killed) != 1 || resp.killed[0] != 300 {
		t.Fatalf("three fuzzy signals must cross the kill threshold: killed=%v", resp.killed)
	}
	contained := false
	for _, r := range inc.recs {
		if r.Action == "contained" {
			contained = true
		}
	}
	if !contained {
		t.Fatalf("three fuzzy signals must record a contained incident: %+v", inc.recs)
	}
}

// TestGuardFuzzy_TrustedProcessIsExempt verifies that a trusted (allowlisted)
// process firing every fuzzy trigger is never scored, mirroring the canary
// exemption.
func TestGuardFuzzy_TrustedProcessIsExempt(t *testing.T) {
	g, resp, inc, tab := newFuzzyGuard(t, entropyReader, false)
	g.deps.Trusted = func(image string) bool { return image == `/opt/backup-agent` }
	tab.Add(proctable.Proc{PID: 500, PPID: 4, Image: `/opt/backup-agent`})

	g.OnFileEvent(FileEvent{PID: 500, Path: `/data/README.txt`, Op: OpWrite})
	for i := 0; i < 12; i++ {
		g.OnFileEvent(FileEvent{PID: 500, Path: "/data/b" + strings.Repeat("x", i) + ".locked", Op: OpWrite})
	}
	if len(resp.killed) != 0 || len(resp.suspended) != 0 || len(inc.recs) != 0 {
		t.Fatalf("trusted process must never escalate: killed=%v suspended=%v recs=%d",
			resp.killed, resp.suspended, len(inc.recs))
	}
}

// TestGuardFuzzy_CanaryShortCircuits verifies that a canary write takes the
// deterministic max-weight branch (immediate kill) and that straggling fuzzy
// events for the same PID cannot re-escalate it.
func TestGuardFuzzy_CanaryShortCircuits(t *testing.T) {
	g, resp, _, _, _, tab := newTestGuard("kill", map[string]bool{`/data/canary.xlsx`: true}, 50, 100)
	tab.Add(proctable.Proc{PID: 300, PPID: 4, Image: `/bin/enc`})

	g.OnFileEvent(FileEvent{PID: 300, Path: `/data/canary.xlsx`, Op: OpWrite})
	if len(resp.killed) != 1 || resp.killed[0] != 300 {
		t.Fatalf("canary write must kill via the deterministic branch: killed=%v", resp.killed)
	}
	// Straggling fuzzy events for the contained PID must not re-fire.
	for i := 0; i < 26; i++ {
		g.OnFileEvent(FileEvent{PID: 300, Path: "/data/m" + strings.Repeat("y", i) + ".log", Op: OpWrite})
	}
	if len(resp.killed) != 1 {
		t.Fatalf("straggling fuzzy events must not re-kill: killed=%v", resp.killed)
	}
}
