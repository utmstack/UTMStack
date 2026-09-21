package ransomware

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
)

type fakeResp struct {
	killed    []int
	suspended []int
	resumed   []int
}

func (f *fakeResp) KillTree(root int, sig string) ([]int, error) { f.killed = append(f.killed, root); return []int{root}, nil }
func (f *fakeResp) Suspend(pid int) error                        { f.suspended = append(f.suspended, pid); return nil }
func (f *fakeResp) Resume(pid int) error                         { f.resumed = append(f.resumed, pid); return nil }

type fakeAppender struct{ lines []string }

func (f *fakeAppender) Append(s string) error { f.lines = append(f.lines, s); return nil }

type fakeQuar struct{ calls int }

func (f *fakeQuar) Quarantine(path, sha, det string) (string, error) { f.calls++; return "q-" + path, nil }

type fakeIncidents struct{ recs []cache.RansomwareIncident }

func (f *fakeIncidents) StoreIncident(r cache.RansomwareIncident) error { f.recs = append(f.recs, r); return nil }

type fakeCanaries struct{ paths map[string]bool }

func (f *fakeCanaries) Contains(p string) bool { return f.paths[p] }

func newTestGuard(mode string, canaries map[string]bool, suspendT, killT int) (*Guard, *fakeResp, *fakeAppender, *fakeQuar, *fakeIncidents, *proctable.Table) {
	tab := proctable.New()
	resp := &fakeResp{}
	app := &fakeAppender{}
	quar := &fakeQuar{}
	inc := &fakeIncidents{}
	cfg := config.Default()
	cfg.Ransomware.Enabled = true
	cfg.Ransomware.ResponseMode = mode
	cfg.Ransomware.SuspendThreshold = suspendT
	cfg.Ransomware.KillThreshold = killT
	g := NewGuard(GuardDeps{
		Cfg: cfg, Table: tab, Resp: resp, Spool: app, Quar: quar, Incidents: inc,
		Canaries:  &fakeCanaries{paths: canaries},
		Hash:      func(string) (string, error) { return "deadbeef", nil },
		Now:       func() time.Time { return time.Unix(1000, 0) },
		NewID:     func() string { return "inc-1" },
	})
	return g, resp, app, quar, inc, tab
}

func TestGuard_T1490_SuspendMode_KillsParentAndQuarantines(t *testing.T) {
	g, resp, app, quar, inc, tab := newTestGuard("suspend", nil, 50, 100)
	// Encryptor (pid 100) spawns vssadmin (pid 200).
	tab.Add(proctable.Proc{PID: 100, PPID: 4, Image: `C:\Users\x\enc.exe`})
	tab.Add(proctable.Proc{PID: 200, PPID: 100, Image: `C:\Windows\System32\vssadmin.exe`})

	g.OnProcStart(200, 100, `C:\Windows\System32\vssadmin.exe`, "vssadmin delete shadows /all /quiet", 1)

	if len(resp.suspended) == 0 || resp.suspended[0] != 100 {
		t.Fatalf("suspend-mode must suspend the culprit (parent 100): %v", resp.suspended)
	}
	if len(resp.killed) == 0 || resp.killed[0] != 100 {
		t.Fatalf("must kill the encryptor parent (100), got %v", resp.killed)
	}
	if quar.calls != 1 {
		t.Fatalf("must quarantine the encryptor image once, got %d", quar.calls)
	}
	if len(inc.recs) != 1 || inc.recs[0].Action != "contained" {
		t.Fatalf("incident not recorded: %+v", inc.recs)
	}
	// A ransomware_contained event was emitted.
	found := false
	for _, l := range app.lines {
		if contains(l, "ransomware_contained") {
			found = true
		}
	}
	if !found {
		t.Fatalf("no ransomware_contained event: %v", app.lines)
	}
}

func TestGuard_Canary_KillMode_NoSuspend(t *testing.T) {
	g, resp, _, quar, _, tab := newTestGuard("kill", map[string]bool{`C:\u\00__a.xlsx`: true}, 50, 100)
	tab.Add(proctable.Proc{PID: 300, PPID: 4, Image: `C:\Users\x\enc.exe`})
	g.OnFileEvent(FileEvent{PID: 300, Path: `C:\u\00__a.xlsx`, Op: OpWrite})

	if len(resp.suspended) != 0 {
		t.Fatalf("kill mode must NOT suspend: %v", resp.suspended)
	}
	if len(resp.killed) == 0 || resp.killed[0] != 300 {
		t.Fatalf("canary tamper must kill toucher 300: %v", resp.killed)
	}
	if quar.calls != 1 {
		t.Fatalf("must quarantine, got %d", quar.calls)
	}
}

func TestGuard_AlertMode_EmitsButNeverActs(t *testing.T) {
	g, resp, app, quar, inc, tab := newTestGuard("alert", map[string]bool{`C:\u\00__a.xlsx`: true}, 50, 100)
	tab.Add(proctable.Proc{PID: 400, PPID: 4, Image: `C:\Users\x\enc.exe`})
	g.OnFileEvent(FileEvent{PID: 400, Path: `C:\u\00__a.xlsx`, Op: OpWrite})

	if len(resp.killed) != 0 || len(resp.suspended) != 0 || quar.calls != 0 {
		t.Fatalf("alert mode must not act: kill=%v susp=%v quar=%d", resp.killed, resp.suspended, quar.calls)
	}
	if len(inc.recs) != 1 || inc.recs[0].Action != "alerted" {
		t.Fatalf("alert incident missing: %+v", inc.recs)
	}
	if len(app.lines) == 0 {
		t.Fatal("alert mode must still emit an event")
	}
}

func TestGuard_Disabled_NoOp(t *testing.T) {
	g, resp, _, quar, inc, tab := newTestGuard("kill", map[string]bool{`C:\u\00__a.xlsx`: true}, 50, 100)
	g.cfg.Ransomware.Enabled = false
	tab.Add(proctable.Proc{PID: 500, PPID: 4, Image: `C:\enc.exe`})
	g.OnFileEvent(FileEvent{PID: 500, Path: `C:\u\00__a.xlsx`, Op: OpWrite})
	if len(resp.killed) != 0 {
		t.Fatal("disabled guard must do nothing")
	}
	if len(resp.suspended) != 0 || quar.calls != 0 || len(inc.recs) != 0 {
		t.Fatalf("disabled guard must do nothing else: susp=%v quar=%d inc=%d", resp.suspended, quar.calls, len(inc.recs))
	}
}

func TestGuard_SuspendMode_SubKill_SuspendsOnly(t *testing.T) {
	// killT 200 so a single weight-100 canary lands on EscSuspend, not EscKill.
	g, resp, _, quar, inc, tab := newTestGuard("suspend", map[string]bool{`C:\u\00__a.xlsx`: true}, 50, 200)
	tab.Add(proctable.Proc{PID: 600, PPID: 4, Image: `C:\Users\x\enc.exe`})
	g.OnFileEvent(FileEvent{PID: 600, Path: `C:\u\00__a.xlsx`, Op: OpWrite})
	if len(resp.suspended) != 1 || resp.suspended[0] != 600 {
		t.Fatalf("suspend mode at EscSuspend must suspend: %v", resp.suspended)
	}
	if len(resp.killed) != 0 || quar.calls != 0 {
		t.Fatalf("EscSuspend must not kill/quarantine: killed=%v quar=%d", resp.killed, quar.calls)
	}
	if len(inc.recs) != 1 || inc.recs[0].Action != "suspended" {
		t.Fatalf("incident action = %+v, want suspended", inc.recs)
	}
}

func TestGuard_KillMode_SubKill_NoSuspend(t *testing.T) {
	// Regression: kill mode must NEVER suspend, even at the EscSuspend rung.
	g, resp, _, quar, inc, tab := newTestGuard("kill", map[string]bool{`C:\u\00__a.xlsx`: true}, 50, 200)
	tab.Add(proctable.Proc{PID: 610, PPID: 4, Image: `C:\Users\x\enc.exe`})
	g.OnFileEvent(FileEvent{PID: 610, Path: `C:\u\00__a.xlsx`, Op: OpWrite})
	if len(resp.suspended) != 0 {
		t.Fatalf("kill mode must not suspend at EscSuspend: %v", resp.suspended)
	}
	if len(resp.killed) != 0 || quar.calls != 0 {
		t.Fatalf("EscSuspend must not kill/quarantine: killed=%v quar=%d", resp.killed, quar.calls)
	}
	if len(inc.recs) != 1 || inc.recs[0].Action != "suspected" {
		t.Fatalf("incident action = %+v, want suspected", inc.recs)
	}
}

func TestGuard_UnknownMode_FailsSafeToAlert(t *testing.T) {
	// Regression: an unknown/mis-cased mode must fail SAFE (alert), never kill.
	g, resp, app, quar, inc, tab := newTestGuard("Monitor", map[string]bool{`C:\u\00__a.xlsx`: true}, 50, 100)
	tab.Add(proctable.Proc{PID: 620, PPID: 4, Image: `C:\Users\x\enc.exe`})
	g.OnFileEvent(FileEvent{PID: 620, Path: `C:\u\00__a.xlsx`, Op: OpWrite})
	if len(resp.killed) != 0 || len(resp.suspended) != 0 || quar.calls != 0 {
		t.Fatalf("unknown mode must not act: killed=%v susp=%v quar=%d", resp.killed, resp.suspended, quar.calls)
	}
	if len(inc.recs) != 1 || inc.recs[0].Action != "alerted" {
		t.Fatalf("unknown mode incident = %+v, want alerted", inc.recs)
	}
	if len(app.lines) == 0 {
		t.Fatal("unknown mode must still emit an event")
	}
}

func TestGuard_Run_SupervisesDeadFeed(t *testing.T) {
	old := feedRetryBackoff
	feedRetryBackoff = time.Millisecond
	defer func() { feedRetryBackoff = old }()

	g, _, _, _, _, _ := newTestGuard("suspend", nil, 50, 100)
	ff := &flakyFeed{}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- g.Run(ctx, ff, 999, func(string) bool { return false }) }()
	time.Sleep(20 * time.Millisecond) // allow several retries
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after ctx cancel")
	}
	if atomic.LoadInt32(&ff.calls) < 2 {
		t.Fatalf("feed was not restarted after it died: calls=%d", ff.calls)
	}
}

type flakyFeed struct{ calls int32 }

func (f *flakyFeed) Run(ctx context.Context, sink func(FileEvent)) error {
	atomic.AddInt32(&f.calls, 1)
	return errors.New("feed died")
}

func contains(s, sub string) bool { return len(s) >= len(sub) && (indexOf(s, sub) >= 0) }
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestGuard_TrustedProcess_NotScored(t *testing.T) {
	tab := proctable.New()
	resp := &fakeResp{}
	cfg := config.Default()
	cfg.Ransomware.Enabled = true
	cfg.Ransomware.ResponseMode = "kill"
	cfg.Ransomware.SuspendThreshold = 50
	cfg.Ransomware.KillThreshold = 100
	trustVeeam := func(image string) bool {
		p := strings.ToLower(strings.ReplaceAll(image, `\`, "/"))
		return strings.HasPrefix(p, "c:/program files/veeam/")
	}
	g := NewGuard(GuardDeps{
		Cfg: cfg, Table: tab, Resp: resp, Spool: &fakeAppender{}, Quar: &fakeQuar{}, Incidents: &fakeIncidents{},
		Canaries: &fakeCanaries{paths: map[string]bool{`C:\decoy\_00.xlsx`: true}},
		Hash:     func(string) (string, error) { return "x", nil },
		Now:      func() time.Time { return time.Unix(1000, 0) },
		NewID:    func() string { return "i" },
		Trusted:  trustVeeam,
	})
	// Trusted backup product (pid 300) runs vssadmin AND touches a canary — both
	// max-weight signals — yet must not be contained.
	tab.Add(proctable.Proc{PID: 300, PPID: 4, Image: `C:\Program Files\Veeam\veeam.exe`})
	tab.Add(proctable.Proc{PID: 301, PPID: 300, Image: `C:\Windows\System32\vssadmin.exe`})
	g.OnProcStart(301, 300, `C:\Windows\System32\vssadmin.exe`, "vssadmin delete shadows /all /quiet", 1)
	g.OnFileEvent(FileEvent{PID: 300, Path: `C:\decoy\_00.xlsx`, Op: OpWrite})
	if len(resp.killed) != 0 || len(resp.suspended) != 0 {
		t.Fatalf("trusted process must not be contained: killed=%v suspended=%v", resp.killed, resp.suspended)
	}

	// Control: an UN-trusted encryptor doing the same IS contained.
	resp2 := &fakeResp{}
	g2 := NewGuard(GuardDeps{
		Cfg: cfg, Table: tab, Resp: resp2, Spool: &fakeAppender{}, Quar: &fakeQuar{}, Incidents: &fakeIncidents{},
		Canaries: &fakeCanaries{paths: map[string]bool{`C:\decoy\_01.xlsx`: true}},
		Hash:     func(string) (string, error) { return "x", nil },
		Now:      func() time.Time { return time.Unix(1000, 0) },
		NewID:    func() string { return "i" }, Trusted: trustVeeam,
	})
	tab.Add(proctable.Proc{PID: 400, PPID: 4, Image: `C:\Users\x\enc.exe`})
	g2.OnFileEvent(FileEvent{PID: 400, Path: `C:\decoy\_01.xlsx`, Op: OpWrite})
	if len(resp2.killed) == 0 {
		t.Fatal("untrusted encryptor tampering a canary should be contained")
	}
}

func TestGuard_SetPolicy_HotAppliesCommandAllowlistAndMode(t *testing.T) {
	g, resp, _, _, _, tab := newTestGuard("kill", nil, 50, 100)
	tab.Add(proctable.Proc{PID: 100, PPID: 4, Image: `C:\Users\x\enc.exe`})
	tab.Add(proctable.Proc{PID: 200, PPID: 100, Image: `C:\Windows\System32\vssadmin.exe`})

	// Hot-apply an allowlist entry that exempts this exact command → no containment.
	g.SetPolicy("kill", []string{"delete shadows"})
	g.OnProcStart(200, 100, `C:\Windows\System32\vssadmin.exe`, "vssadmin delete shadows /all", 1)
	if len(resp.killed) != 0 {
		t.Fatalf("allowlisted command (post-SetPolicy) must not be contained: killed=%v", resp.killed)
	}

	// Clear the allowlist and switch to alert mode → escalation surfaces but no kill.
	g.SetPolicy("alert", nil)
	tab.Add(proctable.Proc{PID: 300, PPID: 4, Image: `C:\Users\x\enc2.exe`})
	tab.Add(proctable.Proc{PID: 301, PPID: 300, Image: `C:\Windows\System32\vssadmin.exe`})
	g.OnProcStart(301, 300, `C:\Windows\System32\vssadmin.exe`, "vssadmin delete shadows /all", 2)
	if len(resp.killed) != 0 || len(resp.suspended) != 0 {
		t.Fatalf("alert mode (post-SetPolicy) must not suspend/kill: killed=%v suspended=%v", resp.killed, resp.suspended)
	}
}
