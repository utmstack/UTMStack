//go:build linux

package responder

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/utmstack/UTMStack/agent/edr/linuxproc"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLinuxResponseSafeHelper(t *testing.T) {
	role := os.Getenv("EDR_RESPONSE_SAFE_HELPER")
	if role == "" {
		return
	}
	if role == "parent" {
		exe, _ := os.Executable()
		child := exec.Command(exe, "-test.run=^TestLinuxResponseSafeHelper$")
		child.Env = append(os.Environ(), "EDR_RESPONSE_SAFE_HELPER=leaf")
		if err := child.Start(); err != nil {
			os.Exit(2)
		}
		fmt.Println(child.Process.Pid)
		defer func() { _ = child.Process.Kill(); _ = child.Wait() }()
	}
	time.Sleep(30 * time.Second)
	os.Exit(0)
}
func linuxSafeChild(t *testing.T, role string) (*exec.Cmd, int64, *bufio.Reader) {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(exe, "-test.run=^TestLinuxResponseSafeHelper$")
	cmd.Env = append(os.Environ(), "EDR_RESPONSE_SAFE_HELPER="+role)
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	id, err := (OSKiller{}).ProcessIdentity(cmd.Process.Pid)
	if err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = (OSKiller{}).KillIdentity(cmd.Process.Pid, id); _ = cmd.Wait() })
	return cmd, id, bufio.NewReader(out)
}
func waitLinuxState(t *testing.T, pid int, stopped bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		p, err := (linuxproc.Reader{}).Read(pid, false)
		if err != nil {
			t.Fatal(err)
		}
		if (p.State == 'T') == stopped {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("process %d state %c, stopped=%v", pid, p.State, stopped)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
func TestLinuxPIDFDSuspendResumeAndKillOwnedChild(t *testing.T) {
	cmd, id, _ := linuxSafeChild(t, "leaf")
	pid := cmd.Process.Pid
	if err := (OSKiller{}).KillIdentity(pid, id+1); !errors.Is(err, ErrProcessGone) {
		t.Fatalf("wrong creation identity accepted: %v", err)
	}
	s := OSSuspender{}
	if err := s.SuspendIdentity(pid, id); errors.Is(err, ErrUnsupported) {
		t.Skip(err)
	} else if err != nil {
		t.Fatal(err)
	}
	defer CloseSuspensions()
	waitLinuxState(t, pid, true)
	if err := s.ResumeIdentity(pid, id+1); !errors.Is(err, ErrProcessGone) {
		t.Fatalf("wrong resume identity accepted: %v", err)
	}
	waitLinuxState(t, pid, true)
	if err := s.ResumeIdentity(pid, id); err != nil {
		t.Fatal(err)
	}
	waitLinuxState(t, pid, false)
	if err := (OSKiller{}).KillIdentity(pid, id); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Wait(); err == nil {
		t.Fatal("owned child survived SIGKILL")
	}
}
func TestLinuxResponseProtectsPIDOneAndSelf(t *testing.T) {
	for _, pid := range []int{0, 1, os.Getpid()} {
		if err := (OSKiller{}).KillIdentity(pid, 1); err == nil {
			t.Fatal("protected PID accepted", pid)
		}
		if err := (OSSuspender{}).SuspendIdentity(pid, 1); err == nil {
			t.Fatal("protected PID suspended", pid)
		}
	}
	if err := (OSSuspender{}).ResumePID(987654321); err == nil {
		t.Fatal("numeric resume without held pidfd accepted")
	}
}
func TestLinuxIdentityBoundTreeKillsOnlyOwnedSafeHelpers(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("full procfs coverage requires dedicated root acceptance")
	}
	cmd, id, out := linuxSafeChild(t, "parent")
	line, err := out.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	child, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil {
		t.Fatal(err)
	}
	childID, err := (OSKiller{}).ProcessIdentity(child)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = (OSKiller{}).KillIdentity(child, childID) }()
	r := New(proctable.New(), nil, OSKiller{}, OSSuspender{})
	if _, err := r.KillTreeIdentity(cmd.Process.Pid, id+1, "Safe.Fixture"); !errors.Is(err, ErrProcessGone) {
		t.Fatalf("reused root not refused: %v", err)
	}
	killed, err := r.KillTreeIdentity(cmd.Process.Pid, id, "Safe.Fixture")
	if errors.Is(err, ErrUnsupported) {
		t.Skip(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	if len(killed) != 2 || killed[0] != child || killed[1] != cmd.Process.Pid {
		t.Fatalf("tree was not leaves-first: %v", killed)
	}
}

// Run permanent registry closure in an isolated process, preserving the rest of
// the package's native acceptance tests regardless of their execution order.
func TestLinuxSuspensionShutdownSealsNewStops(t *testing.T) {
	if os.Getenv("EDR_LINUX_SUSPENSION_CLOSE_TEST") != "1" {
		exe, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, exe, "-test.run=^TestLinuxSuspensionShutdownSealsNewStops$", "-test.v")
		cmd.Env = append(os.Environ(), "EDR_LINUX_SUSPENSION_CLOSE_TEST=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("isolated closure test: %v\n%s", err, out)
		}
		return
	}
	cmd, id, _ := linuxSafeChild(t, "leaf")
	pid := cmd.Process.Pid
	s := OSSuspender{}
	if err := s.SuspendIdentity(pid, id); errors.Is(err, ErrUnsupported) {
		t.Skip(err)
	} else if err != nil {
		t.Fatal(err)
	}
	waitLinuxState(t, pid, true)
	var wg sync.WaitGroup
	results := make(chan error, 32)
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); results <- s.SuspendIdentity(pid, id) }()
	}
	if err := CloseSuspensions(); err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	close(results)
	for err := range results {
		if err != nil && !errors.Is(err, ErrSuspensionsClosed) {
			t.Fatal(err)
		}
	}
	waitLinuxState(t, pid, false)
	if err := s.SuspendIdentity(pid, id); !errors.Is(err, ErrSuspensionsClosed) {
		t.Fatalf("new stop after shutdown accepted: %v", err)
	}
	if err := CloseSuspensions(); err != nil {
		t.Fatalf("second close: %v", err)
	}
	suspendedLinux.Lock()
	defer suspendedLinux.Unlock()
	if !suspendedLinux.closed || len(suspendedLinux.byPID) != 0 {
		t.Fatal("shutdown registry not sealed and drained")
	}
}
