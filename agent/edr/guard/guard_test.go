package guard

import (
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
)

type fakeScanner struct {
	verdict, sig string
	delay        time.Duration
}

func (f fakeScanner) ScanFile(path, source string) (string, string, error) {
	if f.delay > 0 {
		time.Sleep(f.delay)
	}
	return f.verdict, f.sig, nil
}

type fakeResp struct {
	suspended, resumed, killed int
}

func (f *fakeResp) KillTree(pid int, sig string) ([]int, error) { f.killed = pid; return []int{pid}, nil }
func (f *fakeResp) Suspend(pid int) error                       { f.suspended = pid; return nil }
func (f *fakeResp) Resume(pid int) error                        { f.resumed = pid; return nil }

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
