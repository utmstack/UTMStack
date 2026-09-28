//go:build linux

package responder

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestMain lets the test re-execute its own binary as a self-contained helper.
// The live test must not depend on any system binary (e.g. /usr/bin/sleep):
// the EDR's own quarantine removes the image of a malicious process, so a
// system binary seeded as "malicious" in an e2e would be gone by the next run.
func TestMain(m *testing.M) {
	if len(os.Args) > 1 && os.Args[1] == "helper" {
		runHelper(os.Args[2:])
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func runHelper(args []string) {
	if len(args) == 0 {
		return
	}
	switch args[0] {
	case "tree":
		// Root of a 3-process tree: spawn two blocking children, wait for them.
		var kids []*exec.Cmd
		for i := 0; i < 2; i++ {
			c := exec.Command(os.Args[0], "helper", "block")
			if err := c.Start(); err != nil {
				os.Exit(1)
			}
			kids = append(kids, c)
		}
		for _, c := range kids {
			_ = c.Wait()
		}
	case "block":
		select {}
	case "tick":
		// Append one line every 100ms — freeze-observable behavior.
		f, err := os.OpenFile(args[1], os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			os.Exit(1)
		}
		for {
			f.WriteString("tick\n")
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// pidAlive reports whether /proc/<pid> exists.
func pidAlive(pid int) bool {
	_, err := os.Stat("/proc/" + strconv.Itoa(pid))
	return err == nil
}

// waitFor polls fn until it returns true or the deadline passes.
func waitFor(t *testing.T, what string, d time.Duration, fn func() bool) {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

// countLines returns the number of lines in path (0 if unreadable).
func countLines(path string) int {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	return strings.Count(string(b), "\n")
}

func TestLinuxKillFreezeLive(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("live kill/freeze test requires root")
	}
	if testing.Short() {
		t.Skip("skipping live test in -short mode")
	}

	// --- Kill part: spawn a 3-process tree of the test helper itself and
	// terminate it leaves-first. ---
	cmd := exec.Command(os.Args[0], "helper", "tree")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start helper tree: %v", err)
	}
	rootPid := cmd.Process.Pid
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	// 5s: liveDescendants reads every /proc/[0-9]*/stat, which is slow on a
	// loaded host — a tight deadline flakes under load, not from the code.
	waitFor(t, "helper tree to spawn both children", 5*time.Second, func() bool {
		d, ok := liveDescendants(rootPid)
		return ok && len(d) == 3
	})

	desc, ok := liveDescendants(rootPid)
	if !ok {
		t.Fatalf("liveDescendants(%d) = ok=false, want the live tree", rootPid)
	}
	byPos := map[int]int{}
	for i, pid := range desc {
		byPos[pid] = i
	}
	if _, present := byPos[rootPid]; !present {
		t.Fatalf("liveDescendants(%d) = %v, missing the root pid", rootPid, desc)
	}
	// Leaves-first: every child must precede the root.
	for pid, pos := range byPos {
		if pid != rootPid && pos >= byPos[rootPid] {
			t.Fatalf("leaf %d must precede root %d: %v", pid, rootPid, desc)
		}
	}

	// Kill each returned pid; all must be gone from /proc within ~3s. (When
	// the two children die, the tree helper's Wait returns and it may exit on
	// its own — ErrProcessGone is the benign outcome there.)
	k := OSKiller{}
	for _, pid := range desc {
		if err := k.KillPID(pid); err != nil && err != ErrProcessGone {
			t.Fatalf("KillPID(%d): %v", pid, err)
		}
	}
	_ = cmd.Wait() // reap the root helper (already dead) so its /proc entry goes away
	waitFor(t, "all killed pids to disappear from /proc", 3*time.Second, func() bool {
		for _, pid := range desc {
			if pidAlive(pid) {
				return false
			}
		}
		return true
	})

	// --- Freeze part: suspend the tick helper, confirm it stops producing,
	// resume, confirm it continues. The cgroup freezer does not reliably change
	// the /proc state character (a frozen task that was sleeping still reads S),
	// so the test observes behavior instead. ---
	outFile := filepath.Join(t.TempDir(), "ticks")
	tick := exec.Command(os.Args[0], "helper", "tick", outFile)
	if err := tick.Start(); err != nil {
		t.Fatalf("start tick helper: %v", err)
	}
	tickPid := tick.Process.Pid
	defer func() {
		_ = tick.Process.Kill()
		_ = tick.Wait()
	}()

	waitFor(t, "tick helper to write its first line", 3*time.Second, func() bool {
		return countLines(outFile) > 0
	})

	s := OSSuspender{}
	if err := s.SuspendPID(tickPid); err != nil {
		t.Fatalf("SuspendPID(%d): %v", tickPid, err)
	}
	// The kernel freezes asynchronously; give it a moment to take effect
	// before sampling the frozen baseline.
	time.Sleep(300 * time.Millisecond)
	frozenAt := countLines(outFile)
	waitFor(t, "frozen helper to stop", 3*time.Second, func() bool {
		return countLines(outFile) == frozenAt
	})
	// Still frozen: no growth across a second sample.
	time.Sleep(300 * time.Millisecond)
	if got := countLines(outFile); got != frozenAt {
		t.Fatalf("frozen helper grew from %d to %d lines — freeze did not take", frozenAt, got)
	}

	if err := s.ResumePID(tickPid); err != nil {
		t.Fatalf("ResumePID(%d): %v", tickPid, err)
	}
	waitFor(t, "resumed helper to write again", 3*time.Second, func() bool {
		return countLines(outFile) > frozenAt
	})

	_ = tick.Process.Kill()
	_ = tick.Wait()
}
