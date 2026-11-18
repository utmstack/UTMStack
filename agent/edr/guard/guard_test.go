package guard

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
)

type fakeScanner struct {
	verdict, sig string
	delay        time.Duration
	err          error
	resp         *fakeResp // when set, a malicious verdict returns a contain that records into it
}

func (f fakeScanner) ScanImage(path, source string) (string, string, func() string, error) {
	if f.delay > 0 {
		time.Sleep(f.delay)
	}
	var contain func() string
	if f.verdict == "malicious" && f.resp != nil {
		contain = func() string { f.resp.calls = append(f.resp.calls, "contain"); return "quarantined" }
	}
	return f.verdict, f.sig, contain, f.err
}

type fakeResp struct {
	suspended, resumed, killed int
	calls                      []string
}

func (f *fakeResp) KillTree(pid int, sig string) ([]int, error) {
	f.killed = pid
	f.calls = append(f.calls, "kill")
	return []int{pid}, nil
}
func (f *fakeResp) Suspend(pid int) error {
	f.suspended = pid
	f.calls = append(f.calls, "suspend")
	return nil
}
func (f *fakeResp) Resume(pid int) error {
	f.resumed = pid
	f.calls = append(f.calls, "resume")
	return nil
}

func suspendCfg() config.EDRConfig {
	c := config.Default()
	c.SuspendOnLaunch = true
	c.SuspendTimeoutMs = 5000
	return c
}

func TestGuardKillsMaliciousLaunch(t *testing.T) {
	fr := &fakeResp{}
	g := New(suspendCfg(), fakeScanner{verdict: "malicious", sig: "Win.Test.EICAR"}, fr)
	g.OnStart(proctable.Proc{PID: 555, Image: `C:\Users\x\a.exe`})
	if fr.suspended != 555 {
		t.Fatalf("expected suspend of 555, got %d", fr.suspended)
	}
	if fr.killed != 555 {
		t.Fatalf("expected kill of 555, got %d", fr.killed)
	}
}

func TestGuardResumesCleanLaunch(t *testing.T) {
	fr := &fakeResp{}
	g := New(suspendCfg(), fakeScanner{verdict: "clean"}, fr)
	g.OnStart(proctable.Proc{PID: 777, Image: `C:\Users\x\ok.exe`})
	if fr.resumed != 777 || fr.killed != 0 {
		t.Fatalf("expected resume of 777 and no kill, got resumed=%d killed=%d", fr.resumed, fr.killed)
	}
}

// The safe default must NOT suspend, but must still kill a malicious launch.
func TestGuardDefaultDoesNotSuspendButStillKills(t *testing.T) {
	fr := &fakeResp{}
	g := New(config.Default(), fakeScanner{verdict: "malicious", sig: "X"}, fr)
	g.OnStart(proctable.Proc{PID: 42, Image: `C:\Users\x\a.exe`})
	if fr.suspended != 0 {
		t.Fatalf("default must not suspend, but suspended %d", fr.suspended)
	}
	if fr.killed != 42 {
		t.Fatalf("malicious launch must still be killed, got killed=%d", fr.killed)
	}
}

// OS/system processes must never be suspended (avoid freezing the machine),
// but are still killed if malicious.
func TestGuardNeverSuspendsSystemImage(t *testing.T) {
	fr := &fakeResp{}
	g := New(suspendCfg(), fakeScanner{verdict: "malicious", sig: "X"}, fr)
	g.OnStart(proctable.Proc{PID: 4, Image: `C:\Windows\System32\svchost.exe`})
	if fr.suspended != 0 {
		t.Fatalf("system image must not be suspended, but suspended %d", fr.suspended)
	}
	if fr.killed != 4 {
		t.Fatalf("malicious system image must still be killed, got killed=%d", fr.killed)
	}
}

// A slow scan must never hold a process suspended past the bound: it is resumed
// (fail-open) so a scan backlog can't freeze the host.
func TestGuardSuspendTimeoutResumes(t *testing.T) {
	c := suspendCfg()
	c.SuspendTimeoutMs = 40 // resume well before the (200ms) scan completes
	fr := &fakeResp{}
	g := New(c, fakeScanner{verdict: "clean", delay: 200 * time.Millisecond}, fr)
	start := time.Now()
	g.OnStart(proctable.Proc{PID: 999, Image: `C:\Users\x\slow.exe`})
	if fr.resumed != 999 {
		t.Fatalf("slow scan should have resumed 999 on timeout, got resumed=%d", fr.resumed)
	}
	if elapsed := time.Since(start); elapsed < 40*time.Millisecond {
		t.Fatalf("returned too fast (%v) — timeout path not exercised", elapsed)
	}
}

// A confirmed malicious verdict must terminate the process even when the scan
// also reports a secondary failure, such as an event that could not be spooled.
func TestGuardKillsMaliciousVerdictDespiteScanError(t *testing.T) {
	for _, cfg := range []config.EDRConfig{config.Default(), suspendCfg()} {
		fr := &fakeResp{}
		g := New(cfg, fakeScanner{verdict: "malicious", sig: "X", err: errors.New("spool append failed")}, fr)
		g.OnStart(proctable.Proc{PID: 99, Image: `C:\Users\x\a.exe`})
		if fr.killed != 99 {
			t.Fatalf("malicious verdict with a secondary error must still kill, killed=%d", fr.killed)
		}
		if fr.resumed != 0 {
			t.Fatalf("a malicious process must not be resumed, resumed=%d", fr.resumed)
		}
	}
	fr := &fakeResp{}
	g := New(suspendCfg(), fakeScanner{verdict: "unknown", err: errors.New("analysis incomplete")}, fr)
	g.OnStart(proctable.Proc{PID: 98, Image: `C:\Users\x\b.exe`})
	if fr.killed != 0 || fr.resumed != 98 {
		t.Fatalf("a failed scan without a verdict stays fail-open: killed=%d resumed=%d", fr.killed, fr.resumed)
	}
}

// A running image usually cannot be moved, so the image is contained only
// after its tree is killed; containment writes the one detection event. A
// verdict with no response due (clean, or already contained) contains nothing.
func TestGuardContainsImageAfterKill(t *testing.T) {
	for _, tc := range []struct {
		name string
		cfg  config.EDRConfig
		want string
	}{
		{"default", config.Default(), "kill,contain"},
		{"suspend", suspendCfg(), "suspend,kill,contain"},
	} {
		fr := &fakeResp{}
		g := New(tc.cfg, fakeScanner{verdict: "malicious", sig: "X", resp: fr}, fr)
		g.OnStart(proctable.Proc{PID: 7, Image: `C:\Users\x\a.exe`})
		if got := strings.Join(fr.calls, ","); got != tc.want {
			t.Fatalf("%s: calls %q, want %q", tc.name, got, tc.want)
		}
	}
	fr := &fakeResp{}
	g := New(suspendCfg(), fakeScanner{verdict: "clean", resp: fr}, fr)
	g.OnStart(proctable.Proc{PID: 8, Image: `C:\Users\x\ok.exe`})
	if got := strings.Join(fr.calls, ","); got != "suspend,resume" {
		t.Fatalf("clean launch: calls %q", got)
	}
	fr = &fakeResp{}
	g = New(config.Default(), fakeScanner{verdict: "malicious", sig: "X"}, fr)
	g.OnStart(proctable.Proc{PID: 9, Image: `C:\Users\x\gone.exe`})
	if got := strings.Join(fr.calls, ","); got != "kill" {
		t.Fatalf("already contained image: calls %q", got)
	}
}
