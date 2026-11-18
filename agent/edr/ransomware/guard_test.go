package ransomware

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
)

type fakeResp struct {
	mu        sync.Mutex
	killed    []int
	suspended []int
	resumed   []int

	// Optional behaviour knobs (zero values keep the simple "kill works" fake).
	tab        *proctable.Table // when set, killed PIDs are removed like responder.KillTree does
	extra      []int            // descendants reported as killed along with the root
	killNone   bool             // KillTree terminates nothing (e.g. access denied, already gone)
	rootLives  bool             // KillTree terminates only the extra descendants, not the root
	suspendErr error            // Suspend fails with this error
}

func (f *fakeResp) KillTree(root int, sig string) ([]int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.killed = append(f.killed, root)
	if f.killNone {
		return nil, nil
	}
	out := append([]int{}, f.extra...)
	if !f.rootLives {
		out = append(out, root)
	}
	if f.tab != nil {
		for _, pid := range out {
			f.tab.Remove(pid)
		}
	}
	return out, nil
}
func (f *fakeResp) Suspend(pid int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.suspended = append(f.suspended, pid)
	return f.suspendErr
}
func (f *fakeResp) Resume(pid int) error { f.resumed = append(f.resumed, pid); return nil }
func (f *fakeResp) killCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.killed)
}

type fakeAppender struct{ lines []string }

func (f *fakeAppender) Append(s string) error { f.lines = append(f.lines, s); return nil }

type fakeQuar struct {
	calls int
	paths []string
	err   error
}

func (f *fakeQuar) Quarantine(path, sha, det string) (string, error) {
	f.calls++
	f.paths = append(f.paths, path)
	if f.err != nil {
		return "", f.err
	}
	return "q-" + path, nil
}

type fakeIncidents struct{ recs []cache.RansomwareIncident }

func (f *fakeIncidents) StoreIncident(r cache.RansomwareIncident) error {
	f.recs = append(f.recs, r)
	return nil
}

type fakeCanaries struct {
	paths map[string]bool
	ids   map[canaryFileIdentity]bool
}

func (f *fakeCanaries) Contains(p string) bool { return f.paths[p] }
func (f *fakeCanaries) ContainsFileIdentity(device, inode uint64) bool {
	return f.ids[canaryFileIdentity{device, inode}]
}

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
		Canaries: &fakeCanaries{paths: canaries},
		Hash:     func(string) (string, error) { return "deadbeef", nil },
		Now:      func() time.Time { return time.Unix(1000, 0) },
		NewID:    func() string { return "inc-1" },
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

// eventActions decodes the "action" of every spooled event.
func eventActions(t *testing.T, app *fakeAppender) []string {
	t.Helper()
	var out []string
	for _, l := range app.lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(l), &m); err != nil {
			t.Fatalf("spooled line is not JSON: %v: %s", err, l)
		}
		a, _ := m["action"].(string)
		out = append(out, a)
	}
	return out
}

func TestGuard_DecoyOpenOrReadNeverContains(t *testing.T) {
	// Regression (Windows Server 2022 run): Test-Path/Get-Item, GetFileAttributesW
	// and a plain read of a decoy only OPEN it. An open is not tampering and must
	// produce no evidence at all.
	g, resp, app, quar, inc, tab := newTestGuard("suspend", map[string]bool{`C:\00__accounts.xlsx`: true}, 50, 100)
	tab.Add(proctable.Proc{PID: 900, PPID: 4, Image: `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`, StartTS: 1})
	g.OnFileEvent(FileEvent{PID: 900, Path: `C:\00__accounts.xlsx`, Op: OpCreate})
	if len(resp.killed) != 0 || len(resp.suspended) != 0 || quar.calls != 0 || len(inc.recs) != 0 || len(app.lines) != 0 {
		t.Fatalf("an open of a decoy must not act: killed=%v susp=%v quar=%d inc=%d events=%d",
			resp.killed, resp.suspended, quar.calls, len(inc.recs), len(app.lines))
	}
}

func TestGuard_NeverQuarantinesWindowsDirectoryImage(t *testing.T) {
	for _, img := range []string{
		`C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`,
		`c:\WINDOWS\system32\cmd.exe`,
		`\\?\C:\Windows\SysWOW64\rundll32.exe`,
		`C:/Windows/explorer.exe`,
	} {
		g, resp, app, quar, inc, tab := newTestGuard("suspend", map[string]bool{`C:\u\00__a.xlsx`: true}, 50, 100)
		tab.Add(proctable.Proc{PID: 910, PPID: 4, Image: img, StartTS: 1})
		g.OnFileEvent(FileEvent{PID: 910, Path: `C:\u\00__a.xlsx`, Op: OpWrite})
		// Kill/suspend policy is unchanged for OS binaries: a real tamper is stopped.
		if len(resp.killed) != 1 || resp.killed[0] != 910 {
			t.Fatalf("%s: a tampering OS binary must still be killed: %v", img, resp.killed)
		}
		if quar.calls != 0 {
			t.Fatalf("%s: an image under the Windows directory must never be quarantined: %v", img, quar.paths)
		}
		if len(inc.recs) != 1 || inc.recs[0].Action != "contained" || inc.recs[0].QuarantineID != "" {
			t.Fatalf("%s: incident = %+v, want contained with no quarantine id", img, inc.recs)
		}
		if a := eventActions(t, app); len(a) != 1 || a[0] != "ransomware_contained" {
			t.Fatalf("%s: events = %v", img, a)
		}
	}
}

func TestGuard_QuarantinesImagesOutsideWindowsDirectory(t *testing.T) {
	// Boundary: look-alike prefixes are not the Windows directory.
	for _, img := range []string{
		`C:\Windows2\enc.exe`,
		`C:\WindowsApps\enc.exe`,
		`D:\Windows\enc.exe`,
		`C:\Users\x\AppData\Local\Temp\enc.exe`,
	} {
		g, _, _, quar, _, tab := newTestGuard("kill", map[string]bool{`C:\u\00__a.xlsx`: true}, 50, 100)
		tab.Add(proctable.Proc{PID: 920, PPID: 4, Image: img, StartTS: 1})
		g.OnFileEvent(FileEvent{PID: 920, Path: `C:\u\00__a.xlsx`, Op: OpWrite})
		if quar.calls != 1 || quar.paths[0] != img {
			t.Fatalf("%s must be quarantined, got %v", img, quar.paths)
		}
	}
}

func TestGuard_ContainedPIDIsNotReEscalated(t *testing.T) {
	// Regression: after the kill the responder drops the PID from the process
	// table and the scorer forgot it, so every trailing ETW event re-escalated
	// (8 extra ransomware_contained events with empty image and ppid 0).
	g, resp, app, quar, inc, tab := newTestGuard("suspend", map[string]bool{`C:\u\00__a.xlsx`: true, `C:\u\01__b.docx`: true}, 50, 100)
	resp.tab = tab
	tab.Add(proctable.Proc{PID: 700, PPID: 4, Image: `C:\Users\x\enc.exe`, StartTS: 111})
	g.OnFileEvent(FileEvent{PID: 700, Path: `C:\u\00__a.xlsx`, Op: OpWrite})
	for i := 0; i < 8; i++ {
		g.OnFileEvent(FileEvent{PID: 700, Path: `C:\u\01__b.docx`, Op: OpWrite})
	}
	// T1490 evidence attributed to the already-contained culprit is ignored too.
	tab.Add(proctable.Proc{PID: 701, PPID: 700, Image: `C:\Windows\System32\vssadmin.exe`, StartTS: 112})
	g.OnProcStart(701, 700, `C:\Windows\System32\vssadmin.exe`, "vssadmin delete shadows /all /quiet", 0)

	if len(resp.killed) != 1 || quar.calls != 1 || len(inc.recs) != 1 {
		t.Fatalf("contained PID re-escalated: killed=%v quar=%d incidents=%d", resp.killed, quar.calls, len(inc.recs))
	}
	if a := eventActions(t, app); len(a) != 1 || a[0] != "ransomware_contained" {
		t.Fatalf("want exactly one ransomware_contained event, got %v", a)
	}
}

func TestGuard_KilledDescendantsAreNotReEscalated(t *testing.T) {
	g, resp, app, _, _, tab := newTestGuard("kill", map[string]bool{`C:\u\00__a.xlsx`: true}, 50, 100)
	resp.tab = tab
	resp.extra = []int{731}
	tab.Add(proctable.Proc{PID: 730, PPID: 4, Image: `C:\Users\x\enc.exe`, StartTS: 1})
	tab.Add(proctable.Proc{PID: 731, PPID: 730, Image: `C:\Users\x\worker.exe`, StartTS: 2})
	g.OnFileEvent(FileEvent{PID: 730, Path: `C:\u\00__a.xlsx`, Op: OpWrite})
	// A trailing event from the child killed in the same tree.
	g.OnFileEvent(FileEvent{PID: 731, Path: `C:\u\00__a.xlsx`, Op: OpWrite})
	if len(resp.killed) != 1 {
		t.Fatalf("a descendant killed with the tree must not start a second containment: %v", resp.killed)
	}
	if a := eventActions(t, app); len(a) != 1 {
		t.Fatalf("want one event, got %v", a)
	}
}

func TestGuard_RecycledPIDAfterContainmentIsScored(t *testing.T) {
	g, resp, _, _, _, tab := newTestGuard("kill", map[string]bool{`C:\u\00__a.xlsx`: true}, 50, 100)
	resp.tab = tab
	tab.Add(proctable.Proc{PID: 740, PPID: 4, Image: `C:\Users\x\enc.exe`, StartTS: 10})
	g.OnFileEvent(FileEvent{PID: 740, Path: `C:\u\00__a.xlsx`, Op: OpWrite})
	// Windows reuses the PID for a new process (new generation in the table).
	tab.Add(proctable.Proc{PID: 740, PPID: 4, Image: `C:\Users\x\enc2.exe`, StartTS: 20})
	g.OnFileEvent(FileEvent{PID: 740, Path: `C:\u\00__a.xlsx`, Op: OpWrite})
	if len(resp.killed) != 2 {
		t.Fatalf("a new process that reuses a contained PID must be scored: %v", resp.killed)
	}
}

func TestGuard_ContainedMarkExpires(t *testing.T) {
	tab := proctable.New()
	resp := &fakeResp{tab: tab}
	cur := time.Unix(1000, 0)
	cfg := config.Default()
	cfg.Ransomware.Enabled = true
	cfg.Ransomware.ResponseMode = "kill"
	cfg.Ransomware.SuspendThreshold = 50
	cfg.Ransomware.KillThreshold = 100
	g := NewGuard(GuardDeps{
		Cfg: cfg, Table: tab, Resp: resp, Spool: &fakeAppender{}, Quar: &fakeQuar{}, Incidents: &fakeIncidents{},
		Canaries: &fakeCanaries{paths: map[string]bool{`C:\u\00__a.xlsx`: true}},
		Now:      func() time.Time { return cur },
	})
	// Unknown culprit (not yet in the table): only the expiry can end the mark.
	g.OnFileEvent(FileEvent{PID: 750, Path: `C:\u\00__a.xlsx`, Op: OpWrite})
	cur = cur.Add(containedTTL / 2)
	g.OnFileEvent(FileEvent{PID: 750, Path: `C:\u\00__a.xlsx`, Op: OpWrite})
	if len(resp.killed) != 1 {
		t.Fatalf("within the mark lifetime the PID must not re-escalate: %v", resp.killed)
	}
	cur = cur.Add(containedTTL)
	g.OnFileEvent(FileEvent{PID: 750, Path: `C:\u\00__a.xlsx`, Op: OpWrite})
	if len(resp.killed) != 2 {
		t.Fatalf("after the mark expires the PID is scored again: %v", resp.killed)
	}
}

func TestGuard_ConcurrentEvidenceContainsOnce(t *testing.T) {
	// The ETW feed and the process dispatch call the guard from separate goroutines.
	g, resp, _, _, _, tab := newTestGuard("kill", map[string]bool{`C:\u\00__a.xlsx`: true}, 50, 100)
	resp.tab = tab
	tab.Add(proctable.Proc{PID: 760, PPID: 4, Image: `C:\Users\x\enc.exe`, StartTS: 1})
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			g.OnFileEvent(FileEvent{PID: 760, Path: `C:\u\00__a.xlsx`, Op: OpWrite})
		}()
	}
	wg.Wait()
	if n := resp.killCount(); n != 1 {
		t.Fatalf("concurrent evidence must contain the PID once, got %d kills", n)
	}
}

func TestGuard_QuarantineFailureEmitsContainFailed(t *testing.T) {
	g, _, app, quar, inc, tab := newTestGuard("suspend", map[string]bool{`C:\u\00__a.xlsx`: true}, 50, 100)
	quar.err = errors.New("Access is denied.")
	tab.Add(proctable.Proc{PID: 770, PPID: 4, Image: `C:\Users\x\enc.exe`, StartTS: 1})
	g.OnFileEvent(FileEvent{PID: 770, Path: `C:\u\00__a.xlsx`, Op: OpWrite})
	if len(inc.recs) != 1 || inc.recs[0].Action != "contain_failed" {
		t.Fatalf("incident = %+v, want contain_failed", inc.recs)
	}
	if a := eventActions(t, app); len(a) != 1 || a[0] != "ransomware_contain_failed" {
		t.Fatalf("the event must report the failed containment, got %v", a)
	}
}

func TestGuard_NothingKilledEmitsContainFailed(t *testing.T) {
	g, resp, app, _, inc, tab := newTestGuard("kill", map[string]bool{`C:\u\00__a.xlsx`: true}, 50, 100)
	resp.killNone = true
	tab.Add(proctable.Proc{PID: 780, PPID: 4, Image: `C:\Users\x\enc.exe`, StartTS: 1})
	g.OnFileEvent(FileEvent{PID: 780, Path: `C:\u\00__a.xlsx`, Op: OpWrite})
	if len(inc.recs) != 1 || inc.recs[0].Action != "contain_failed" {
		t.Fatalf("incident = %+v, want contain_failed", inc.recs)
	}
	if a := eventActions(t, app); len(a) != 1 || a[0] != "ransomware_contain_failed" {
		t.Fatalf("the event must report the failed containment, got %v", a)
	}
}

func TestGuard_RootNotKilledEmitsContainFailed(t *testing.T) {
	// Killing only a child (e.g. conhost.exe) while the culprit itself survives
	// (access denied) leaves the encryptor running: that is not containment.
	g, resp, app, _, inc, tab := newTestGuard("kill", map[string]bool{`C:\u\00__a.xlsx`: true}, 50, 100)
	resp.extra = []int{786}
	resp.rootLives = true
	tab.Add(proctable.Proc{PID: 785, PPID: 4, Image: `C:\Users\x\enc.exe`, StartTS: 1})
	tab.Add(proctable.Proc{PID: 786, PPID: 785, Image: `C:\Windows\System32\conhost.exe`, StartTS: 2})
	g.OnFileEvent(FileEvent{PID: 785, Path: `C:\u\00__a.xlsx`, Op: OpWrite})
	if len(inc.recs) != 1 || inc.recs[0].Action != "contain_failed" {
		t.Fatalf("incident = %+v, want contain_failed", inc.recs)
	}
	if a := eventActions(t, app); len(a) != 1 || a[0] != "ransomware_contain_failed" {
		t.Fatalf("the event must report the failed containment, got %v", a)
	}
}

func TestGuard_SuspendFailureIsRecorded(t *testing.T) {
	g, resp, app, _, inc, tab := newTestGuard("suspend", map[string]bool{`C:\u\00__a.xlsx`: true}, 50, 200)
	resp.suspendErr = errors.New("Access is denied.")
	tab.Add(proctable.Proc{PID: 790, PPID: 4, Image: `C:\Users\x\enc.exe`, StartTS: 1})
	g.OnFileEvent(FileEvent{PID: 790, Path: `C:\u\00__a.xlsx`, Op: OpWrite})
	if len(inc.recs) != 1 || inc.recs[0].Action != "suspend_failed" {
		t.Fatalf("incident = %+v, want suspend_failed", inc.recs)
	}
	if a := eventActions(t, app); len(a) != 1 || a[0] != "ransomware_suspected" {
		t.Fatalf("events = %v, want one ransomware_suspected", a)
	}
}

type interestFeed struct {
	isDecoy func(string) bool
	got     chan func(string) bool
}

func (f *interestFeed) setInterest(isDecoy func(string) bool) { f.isDecoy = isDecoy }
func (f *interestFeed) Run(ctx context.Context, sink func(FileEvent)) error {
	f.got <- f.isDecoy
	<-ctx.Done()
	return ctx.Err()
}

func TestGuard_Run_GivesFeedTheDecoySet(t *testing.T) {
	// The ETW feed must know the decoy paths to correlate Write/SetInformation
	// events (which carry no file name) back to a decoy open.
	g, _, _, _, _, _ := newTestGuard("suspend", map[string]bool{`C:\u\00__a.xlsx`: true}, 50, 100)
	ff := &interestFeed{got: make(chan func(string) bool, 1)}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = g.Run(ctx, ff, 999, nil) }()
	select {
	case isDecoy := <-ff.got:
		if isDecoy == nil || !isDecoy(`C:\u\00__a.xlsx`) || isDecoy(`C:\u\other.xlsx`) {
			t.Fatal("feed did not receive the guard's decoy set")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("feed never ran")
	}
}

// A process that tampers within milliseconds of starting is not in the
// WMI-fed table yet; its image is then read from the live process so it can
// still be quarantined and reported.
func TestGuard_UnknownImageResolvedFromLiveProcess(t *testing.T) {
	decoy := `C:\00__accounts.xlsx`
	g, resp, app, quar, _, _ := newTestGuard("suspend", map[string]bool{decoy: true}, 50, 100)
	g.deps.LiveProcess = func(pid int) (string, int, error) {
		if pid != 500 {
			t.Fatalf("live lookup for unexpected pid %d", pid)
		}
		return `C:\Users\x\fast.exe`, 77, nil
	}
	g.OnFileEvent(FileEvent{PID: 500, Path: decoy, Op: OpWrite})
	if len(resp.killed) != 1 || resp.killed[0] != 500 {
		t.Fatalf("killed = %v", resp.killed)
	}
	if len(quar.paths) != 1 || quar.paths[0] != `C:\Users\x\fast.exe` {
		t.Fatalf("quarantined = %v", quar.paths)
	}
	last := app.lines[len(app.lines)-1]
	if !strings.Contains(last, `fast.exe`) || !strings.Contains(last, `"ppid":77`) || !strings.Contains(last, "ransomware_contained") {
		t.Fatalf("event lacks the resolved process: %s", last)
	}
}

// The allowlist must also recognize a trusted program the table has not seen.
func TestGuard_UnknownImageTrustedViaLiveProcess(t *testing.T) {
	decoy := `C:\00__accounts.xlsx`
	g, resp, app, quar, _, _ := newTestGuard("suspend", map[string]bool{decoy: true}, 50, 100)
	g.deps.LiveProcess = func(int) (string, int, error) { return `C:\Program Files\Backup\backup.exe`, 4, nil }
	g.deps.Trusted = func(image string) bool { return strings.HasSuffix(strings.ToLower(image), `\backup.exe`) }
	g.OnFileEvent(FileEvent{PID: 600, Path: decoy, Op: OpDelete})
	if len(resp.killed) != 0 || quar.calls != 0 || len(app.lines) != 0 {
		t.Fatalf("trusted process scored: killed=%v quarantine=%d events=%d", resp.killed, quar.calls, len(app.lines))
	}
}

// When the live lookup fails too, the process is still stopped; nothing is
// quarantined because no image is known.
func TestGuard_LiveProcessLookupFailureStillStops(t *testing.T) {
	decoy := `C:\00__accounts.xlsx`
	g, resp, _, quar, _, _ := newTestGuard("suspend", map[string]bool{decoy: true}, 50, 100)
	g.deps.LiveProcess = func(int) (string, int, error) { return "", 0, errors.New("gone") }
	g.OnFileEvent(FileEvent{PID: 700, Path: decoy, Op: OpRename})
	if len(resp.killed) != 1 || quar.calls != 0 {
		t.Fatalf("killed=%v quarantine calls=%d", resp.killed, quar.calls)
	}
}
