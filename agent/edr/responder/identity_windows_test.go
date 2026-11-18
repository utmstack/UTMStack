//go:build windows

package responder

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"
)

func TestWindowsIdentitySafeFixture(t *testing.T) {
	if os.Getenv("EDR_WINDOWS_RESPONSE_SAFE_HELPER") == "1" {
		time.Sleep(30 * time.Second)
		os.Exit(0)
	}
}
func TestWindowsIdentityBoundSuspendAndKillSafeChild(t *testing.T) {
	exe, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	cmd := exec.Command(exe, "-test.run=^TestWindowsIdentitySafeFixture$")
	cmd.Env = append(os.Environ(), "EDR_WINDOWS_RESPONSE_SAFE_HELPER=1")
	if e = cmd.Start(); e != nil {
		t.Fatal(e)
	}
	defer func() { _ = cmd.Process.Kill(); _ = cmd.Wait(); _ = CloseSuspensions() }()
	pid := cmd.Process.Pid
	id, e := (OSKiller{}).ProcessIdentity(pid)
	if e != nil {
		t.Fatal(e)
	}
	p, e := processObservation(pid)
	if e != nil {
		t.Fatal(e)
	}
	if p.identity != id || p.ppid != os.Getpid() {
		t.Fatalf("observed wrong child: %+v", p)
	}
	if e = (OSKiller{}).KillIdentity(pid, id+1); !errors.Is(e, ErrProcessGone) {
		t.Fatalf("wrong identity accepted: %v", e)
	}
	s := OSSuspender{}
	if e = s.SuspendIdentity(pid, id); e != nil {
		t.Fatal(e)
	}
	if e = s.ResumeIdentity(pid, id+1); !errors.Is(e, ErrProcessGone) {
		t.Fatalf("wrong resume identity: %v", e)
	}
	if e = s.ResumeIdentity(pid, id); e != nil {
		t.Fatal(e)
	}
	if e = (OSKiller{}).KillIdentity(pid, id); e != nil {
		t.Fatal(e)
	}
	if e = cmd.Wait(); e == nil {
		t.Fatal("owned helper survived termination")
	}
}
func TestWindowsIdentityResponseProtectsCorePIDsAndNumericResume(t *testing.T) {
	for _, pid := range []int{0, 1, 4, os.Getpid()} {
		if e := (OSKiller{}).KillIdentity(pid, 1); e == nil {
			t.Fatalf("protected PID %d accepted", pid)
		}
	}
	if e := (OSSuspender{}).ResumePID(987654321); e == nil {
		t.Fatal("numeric resume without original handle accepted")
	}
}

func TestWindowsSuspensionShutdownSealsNewStops(t *testing.T) {
	if os.Getenv("EDR_WINDOWS_SUSPENSION_CLOSE_TEST") != "1" {
		exe, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, exe, "-test.run=^TestWindowsSuspensionShutdownSealsNewStops$", "-test.v")
		cmd.Env = append(os.Environ(), "EDR_WINDOWS_SUSPENSION_CLOSE_TEST=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("isolated closure test: %v\n%s", err, out)
		}
		return
	}
	exe, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	cmd := exec.Command(exe, "-test.run=^TestWindowsIdentitySafeFixture$")
	cmd.Env = append(os.Environ(), "EDR_WINDOWS_RESPONSE_SAFE_HELPER=1")
	if e = cmd.Start(); e != nil {
		t.Fatal(e)
	}
	defer func() { _ = cmd.Process.Kill(); _ = cmd.Wait(); _ = CloseSuspensions() }()
	pid := cmd.Process.Pid
	id, e := (OSKiller{}).ProcessIdentity(pid)
	if e != nil {
		t.Fatal(e)
	}
	s := OSSuspender{}
	if e = s.SuspendIdentity(pid, id); e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	results := make(chan error, 32)
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); results <- s.SuspendIdentity(pid, id) }()
	}
	if e = CloseSuspensions(); e != nil {
		t.Fatal(e)
	}
	wg.Wait()
	close(results)
	for e := range results {
		if e != nil && !errors.Is(e, ErrSuspensionsClosed) {
			t.Fatal(e)
		}
	}
	if e = s.SuspendIdentity(pid, id); !errors.Is(e, ErrSuspensionsClosed) {
		t.Fatalf("new stop after shutdown accepted: %v", e)
	}
	if e = CloseSuspensions(); e != nil {
		t.Fatal(e)
	}
	suspendedWindows.Lock()
	defer suspendedWindows.Unlock()
	if !suspendedWindows.closed || len(suspendedWindows.byPID) != 0 {
		t.Fatal("shutdown registry not sealed and drained")
	}
}
