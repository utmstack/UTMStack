//go:build linux

package procwatch

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

type processChannel struct{ events chan ProcStart }

func (h processChannel) OnProcStart(p ProcStart) { h.events <- p }
func TestLinuxProcessFixtureChild(t *testing.T) {
	if os.Getenv("EDR_PROCESS_FIXTURE_SHORT") == "1" {
		time.Sleep(100 * time.Millisecond)
		os.Exit(0)
	}
	if os.Getenv("EDR_PROCESS_FIXTURE_CHILD") == "1" {
		time.Sleep(30 * time.Second)
		os.Exit(0)
	}
}

func TestLinuxConnectorObservesOwnedShortLivedImage(t *testing.T) {
	if os.Getenv("OPENEDR_ISOLATED_VM") != "1" {
		t.Skip("requires explicit isolated native VM acceptance")
	}
	h := processChannel{make(chan ProcStart, 8192)}
	w := NewWithMode(h, "connector")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- w.RunError(ctx) }()
	deadline := time.Now().Add(5 * time.Second)
	for w.Mode() != "connector" {
		select {
		case err := <-done:
			t.Fatalf("connector startup: %v", err)
		default:
		}
		if time.Now().After(deadline) {
			t.Fatal("connector not ready", w.Health())
		}
		time.Sleep(time.Millisecond)
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	child := exec.Command(exe, "-test.run=^TestLinuxProcessFixtureChild$")
	child.Env = append(os.Environ(), "EDR_PROCESS_FIXTURE_SHORT=1")
	if err = child.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = child.Process.Kill(); _ = child.Wait() }()
	inventoriedSelf := false
	for {
		select {
		case p := <-h.events:
			if p.PID == os.Getpid() && p.EventKind == "inventory" && p.OSIdentity > 0 {
				inventoriedSelf = true
			}
			if p.PID == child.Process.Pid && (p.EventKind == "exec" || p.EventKind == "snapshot") && strings.Contains(p.Cmdline, "^TestLinuxProcessFixtureChild$") {
				if !inventoriedSelf {
					t.Fatal("existing parent was not inventoried before new process")
				}
				if p.Image == "" || p.OSIdentity <= 0 || p.PPID != os.Getpid() {
					t.Fatalf("unbound observation %+v", p)
				}
				cancel()
				select {
				case <-done:
					return
				case <-time.After(time.Second):
					t.Fatal("connector failed to cancel")
				}
			}
		case err := <-done:
			t.Fatalf("connector: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("short-lived owned exec not observed")
		}
	}
}
func TestLinuxProcPollingObservesOnlyOwnedSafeChild(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("full procfs coverage requires dedicated root acceptance")
	}
	h := processChannel{make(chan ProcStart, 1024)}
	w := NewWithMode(h, "poll")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- w.RunError(ctx) }()
	deadline := time.Now().Add(5 * time.Second)
	for w.Health() != nil {
		select {
		case err := <-done:
			t.Fatalf("procfs unavailable (run root acceptance): %v", err)
		default:
		}
		if time.Now().After(deadline) {
			t.Fatal(w.Health())
		}
		time.Sleep(10 * time.Millisecond)
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	child := exec.Command(exe, "-test.run=^TestLinuxProcessFixtureChild$")
	child.Env = append(os.Environ(), "EDR_PROCESS_FIXTURE_CHILD=1")
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = child.Process.Kill(); _ = child.Wait() }()
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	found := false
	for !found {
		select {
		case p := <-h.events:
			if p.PID == child.Process.Pid {
				if p.OSIdentity <= 0 || p.PPID != os.Getpid() || p.Image == "" || p.EventKind != "snapshot" {
					t.Fatalf("incomplete observation %+v", p)
				}
				found = true
			}
		case err := <-done:
			t.Fatal(err)
		case <-timer.C:
			t.Fatal("safe child not observed")
		}
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("process monitor did not cancel")
	}
}
