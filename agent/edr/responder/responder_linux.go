//go:build linux

package responder

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/shared/logger"
	"golang.org/x/sys/unix"
)

// cgroup v2 freezer layout: one leaf cgroup per frozen pid, so a freeze is
// always a single-process operation and cleanup is a plain directory remove.
const (
	cgroupRoot      = "/sys/fs/cgroup"
	freezeDir       = cgroupRoot + "/utmstack_edr/freeze"
	freezePollEvery = 50 * time.Millisecond
	freezePollLimit = 2 * time.Second
)

// cgroupV2Active reports whether /sys/fs/cgroup is a unified (cgroup v2)
// hierarchy. v1 layouts and absent mounts both fail open: freeze is refused
// (logged once) while kill + tree walk keep working on signals + /proc.
func cgroupV2Active() bool {
	fi, err := os.Stat(cgroupRoot)
	if err != nil || !fi.IsDir() {
		return false
	}
	_, err = os.Stat(filepath.Join(cgroupRoot, "cgroup.controllers"))
	return err == nil
}

// freezeWarnOnce keeps a cgroup-v1 (or absent) host from getting a warning
// per suspend attempt — one line is enough to explain why freeze is unavailable.
var freezeWarnOnce sync.Once

func warnFreezeUnavailable() {
	freezeWarnOnce.Do(func() {
		logger.Debug(logger.LevelWarning, "UTMStack EDR: cgroup v2 not active at %s — process freeze unavailable; kill and tree walk still work", cgroupRoot)
	})
}

// OSKiller terminates a process with SIGKILL.
type OSKiller struct{}

func (OSKiller) KillPID(pid int) error {
	err := unix.Kill(pid, unix.SIGKILL)
	if err == unix.ESRCH {
		// The PID no longer exists — already terminated (benign tree-kill race).
		return ErrProcessGone
	}
	if err != nil {
		return fmt.Errorf("kill pid %d: %w", pid, err)
	}
	return nil
}

// OSSuspender freezes/unfreezes a single process via its own cgroup v2 leaf.
type OSSuspender struct{}

func (OSSuspender) SuspendPID(pid int) error {
	if !cgroupV2Active() {
		warnFreezeUnavailable()
		return fmt.Errorf("cgroup v2 freezer unavailable")
	}
	leaf := filepath.Join(freezeDir, strconv.Itoa(pid))
	if err := os.MkdirAll(leaf, 0o755); err != nil {
		return fmt.Errorf("create freeze cgroup for pid %d: %w", pid, err)
	}
	if err := writeFileAtomic(filepath.Join(leaf, "cgroup.procs"), strconv.Itoa(pid)); err != nil {
		_ = os.Remove(leaf)
		return fmt.Errorf("attach pid %d to freeze cgroup: %w", pid, err)
	}
	freezeFile := filepath.Join(leaf, "cgroup.freeze")
	if err := writeFileAtomic(freezeFile, "1"); err != nil {
		_ = os.Remove(leaf)
		return fmt.Errorf("freeze pid %d: %w", pid, err)
	}
	// The kernel freezes the cgroup asynchronously after the write; confirm it
	// actually took (bounded). SuspendPID returns nil only when the freeze is
	// confirmed, so callers can rely on the process being stopped.
	deadline := time.Now().Add(freezePollLimit)
	for {
		if b, err := os.ReadFile(freezeFile); err == nil && strings.TrimSpace(string(b)) == "1" {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("freeze pid %d: cgroup.freeze not confirmed within %s", pid, freezePollLimit)
		}
		time.Sleep(freezePollEvery)
	}
}

func (OSSuspender) ResumePID(pid int) error {
	leaf := filepath.Join(freezeDir, strconv.Itoa(pid))
	freezeFile := filepath.Join(leaf, "cgroup.freeze")
	if _, err := os.Stat(freezeFile); err != nil {
		// No leaf cgroup for this pid — it was never frozen (or already
		// released), so there is nothing to do.
		return nil
	}
	if err := writeFileAtomic(freezeFile, "0"); err != nil {
		return fmt.Errorf("unfreeze pid %d: %w", pid, err)
	}
	// Poll until the kernel confirms the cgroup is fully unfrozen (bounded:
	// never block the caller on a stuck freeze).
	deadline := time.Now().Add(freezePollLimit)
	for {
		b, err := os.ReadFile(freezeFile)
		if err == nil && strings.TrimSpace(string(b)) == "0" {
			break
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("unfreeze pid %d: cgroup.freeze did not clear within %s", pid, freezePollLimit)
		}
		time.Sleep(freezePollEvery)
	}
	removeFreezeDirs(leaf)
	return nil
}

// writeFileAtomic writes a single short value to a cgroup file.
func writeFileAtomic(path, value string) error {
	return os.WriteFile(path, []byte(value), 0o200)
}

// removeFreezeDirs removes the leaf cgroup and then best-effort removes the
// now-possibly-empty parents, stopping at the utmstack_edr directory (never
// the cgroup mount point itself). ENOTEMPTY/ENOENT are expected and ignored —
// other pids may still be frozen, or the dirs are gone.
func removeFreezeDirs(leaf string) {
	_ = os.Remove(leaf)
	for _, dir := range []string{filepath.Dir(leaf), filepath.Dir(filepath.Dir(leaf))} {
		if filepath.Base(dir) == "utmstack_edr" {
			break
		}
		if err := os.Remove(dir); err != nil {
			if !os.IsNotExist(err) && err != unix.ENOTEMPTY {
				logger.Debug(logger.LevelWarning, "UTMStack EDR: freezer cleanup: remove %s: %v", dir, err)
			}
		}
	}
}

// ReleaseAllFreezes unfreezes and removes every /sys/fs/cgroup/utmstack_edr/
// freeze/<pid> entry. Called from service Stop and at pipeline start (clearing
// stale freezes left by a crashed daemon run). Fail-open: it never panics and
// never blocks; errors are logged.
func ReleaseAllFreezes() {
	if !cgroupV2Active() {
		return
	}
	entries, err := os.ReadDir(freezeDir)
	if err != nil {
		if !os.IsNotExist(err) {
			logger.Debug(logger.LevelWarning, "UTMStack EDR: release freezes: read %s: %v", freezeDir, err)
		}
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		leaf := filepath.Join(freezeDir, e.Name())
		if b, err := os.ReadFile(filepath.Join(leaf, "cgroup.freeze")); err == nil && strings.TrimSpace(string(b)) == "1" {
			_ = writeFileAtomic(filepath.Join(leaf, "cgroup.freeze"), "0")
		}
		_ = os.Remove(leaf)
	}
	removeFreezeDirs(freezeDir)
}

// parseStatLine extracts the pid and ppid from one /proc/<pid>/stat line. The
// ppid is field 4 AFTER the last ')' — the comm field (2) may itself contain
// spaces and parentheses, so splitting on the first '(' would misparse it.
func parseStatLine(line string) (pid, ppid int, ok bool) {
	open := strings.IndexByte(line, '(')
	close := strings.LastIndexByte(line, ')')
	if open < 0 || close < open || close+1 >= len(line) {
		return 0, 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(line[:open]))
	if err != nil {
		return 0, 0, false
	}
	fields := strings.Fields(line[close+1:])
	if len(fields) < 2 {
		return 0, 0, false
	}
	ppid, err = strconv.Atoi(fields[1]) // field 4 overall: state, ppid, ...
	if err != nil {
		return 0, 0, false
	}
	return pid, ppid, true
}

// liveDescendants returns rootPID and all of its transitive children from a
// live /proc snapshot, ordered leaves-first (post-order DFS). This catches
// children the async proctable feed has not registered yet — the common
// tree-kill race where a process spawns a helper microseconds before it is
// terminated. Kernel threads (ppid 0 or 2) are excluded, self-parent cycles
// from PID reuse are guarded, and ok=false (root not in the snapshot) makes
// the caller fall back to the proctable.
func liveDescendants(rootPID int) ([]int, bool) {
	return descendantsFrom(procSnapshot(), rootPID)
}

// procSnapshot reads /proc/[0-9]*/stat into a pid→ppid map.
func procSnapshot() map[int]int {
	procs, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}
	snap := make(map[int]int, len(procs))
	for _, e := range procs {
		if !e.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		b, err := os.ReadFile(filepath.Join("/proc", e.Name(), "stat"))
		if err != nil {
			continue // process exited between the listing and the read
		}
		if _, ppid, ok := parseStatLine(string(b)); ok {
			snap[pid] = ppid
		}
	}
	return snap
}

// descendantsFrom walks the pid→ppid snapshot from rootPID, post-order (leaves
// first), including rootPID itself. Extracted from liveDescendants so the
// tree logic is testable against synthetic snapshots.
func descendantsFrom(snap map[int]int, rootPID int) ([]int, bool) {
	if _, exists := snap[rootPID]; !exists {
		return nil, false
	}
	children := make(map[int][]int, len(snap))
	for pid, ppid := range snap {
		if ppid == 0 || ppid == 2 { // kernel threads are not part of any user tree
			continue
		}
		children[ppid] = append(children[ppid], pid)
	}
	var order []int
	seen := make(map[int]bool, len(snap))
	var visit func(int)
	visit = func(id int) {
		if seen[id] {
			return
		}
		seen[id] = true
		for _, c := range children[id] {
			if c != id { // guard against a self-parent cycle from PID reuse
				visit(c)
			}
		}
		order = append(order, id)
	}
	visit(rootPID)
	return order, true
}
