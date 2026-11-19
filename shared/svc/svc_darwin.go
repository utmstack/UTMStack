//go:build darwin

package svc

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// On macOS services are managed by launchd, not systemd, so none of the
// systemctl calls in svc_unix.go (linux) work here. The agent and updater are
// installed by kardianos/service as system launch daemons (optionUserService
// defaults to false): the plist lives at /Library/LaunchDaemons/<name>.plist
// with the service name as its Label. launchctl load/unload take that plist
// PATH; `launchctl list <label>` takes the bare label.
//
// Before this file existed, the darwin build compiled the `linux || darwin`
// svc_unix.go, which called systemctl on every self-update and failed with
// exec: "systemctl": executable file not found in $PATH - so the macOS agent
// could never update itself.
//
// Not-loaded detection: `launchctl list <label>` exits 113 (no output on
// macOS 12+, "Could not find service ..." on older releases) when the service
// is not loaded. That exit code is the reliable signal; a non-zero exit that
// is NOT not-found (permission, transient launchd failure) must be surfaced,
// never treated as "stopped" - the updater swaps the agent binary after Stop,
// so a false "stopped" would replace a running executable.

// darwinStopRetry lets launchd finish tearing the daemon down before Restart
// reloads it.
const darwinStopRetry = 50 * time.Millisecond

// runDarwin executes a command, returning its combined output, the process
// exit code, and an error (non-nil iff the process exited non-zero, matching
// exec.Cmd.CombinedOutput semantics). If the binary itself cannot be executed
// the exit code is -1. Overridden in tests.
var runDarwin = func(name string, args ...string) (string, int, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		exitCode = -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}
	return string(out), exitCode, err
}

// plistInstalled reports whether the service's launch daemon plist exists.
// Overridden in tests.
var plistInstalled = func(serviceName string) bool {
	_, err := os.Stat(daemonPlistPath(serviceName))
	return err == nil
}

// Start starts a launch daemon by loading its plist. launchctl load is NOT
// idempotent - loading an already-loaded daemon fails with exit 37
// ("Operation already in progress") or 17 ("File exists") - so Start first
// checks `launchctl list` and is a no-op when the daemon is already loaded.
func Start(serviceName string) error {
	if !validLaunchdName(serviceName) {
		return fmt.Errorf("invalid service name %q", serviceName)
	}
	if _, exit, _ := runDarwin("launchctl", "list", serviceName); exit == 0 {
		return nil // already loaded: nothing to do
	}
	plistPath := daemonPlistPath(serviceName)
	if !plistInstalled(serviceName) {
		return fmt.Errorf("service %s is not installed: %s", serviceName, plistPath)
	}
	if out, _, err := runDarwin("launchctl", "load", plistPath); err != nil {
		return fmt.Errorf("failed to start service %s: %v (%s)", serviceName, err, strings.TrimSpace(out))
	}
	return nil
}

// Stop stops a launch daemon by unloading its plist. Unloading a daemon that
// is not currently loaded fails; that is a successful no-op (the updater
// stops the agent before swapping the binary, which may already be down).
// Only the genuine "not loaded" result is treated as a no-op - if the
// follow-up query fails for another reason, or the daemon still shows a PID,
// Stop surfaces an error so the update does not proceed over a running agent.
func Stop(serviceName string) error {
	if !validLaunchdName(serviceName) {
		return fmt.Errorf("invalid service name %q", serviceName)
	}
	plistPath := daemonPlistPath(serviceName)
	if !plistInstalled(serviceName) {
		return fmt.Errorf("service %s is not installed: %s", serviceName, plistPath)
	}
	out, _, err := runDarwin("launchctl", "unload", plistPath)
	if err == nil {
		return nil
	}
	lout, lexit, lerr := runDarwin("launchctl", "list", serviceName)
	switch {
	case lerr != nil && launchctlListNotFound(lexit, lout):
		return nil // daemon is genuinely not loaded: benign no-op
	case lerr == nil && !plistRunningPID(lout):
		return nil // registered but not running: safe to proceed
	default:
		return fmt.Errorf("failed to stop service %s: unload: %v (%s)", serviceName, err, strings.TrimSpace(out))
	}
}

// Restart restarts a launch daemon by unloading then loading its plist.
func Restart(serviceName string) error {
	if err := Stop(serviceName); err != nil {
		return err
	}
	time.Sleep(darwinStopRetry)
	return Start(serviceName)
}

// IsActive checks if a launch daemon is running. A daemon that is not loaded
// is not running, without error (the updater health-checks right after Start);
// any other launchctl failure is propagated rather than misreported as idle.
func IsActive(serviceName string) (bool, error) {
	if !validLaunchdName(serviceName) {
		return false, fmt.Errorf("invalid service name %q", serviceName)
	}
	out, exit, err := runDarwin("launchctl", "list", serviceName)
	if err != nil {
		if launchctlListNotFound(exit, out) {
			return false, nil
		}
		return false, fmt.Errorf("failed to query service %s: %v (%s)", serviceName, err, strings.TrimSpace(out))
	}
	return plistRunningPID(out), nil
}

// Status returns the status of a launch daemon: running (PID present),
// stopped (loaded and idle, or installed but not loaded), unknown (not
// installed, or the query failed for a non-not-found reason).
func Status(serviceName string) (string, error) {
	if !validLaunchdName(serviceName) {
		return StatusUnknown, fmt.Errorf("invalid service name %q", serviceName)
	}
	if !plistInstalled(serviceName) {
		return StatusUnknown, nil
	}
	out, exit, err := runDarwin("launchctl", "list", serviceName)
	if err != nil {
		if launchctlListNotFound(exit, out) {
			return StatusStopped, nil
		}
		return StatusUnknown, fmt.Errorf("failed to query service %s: %v (%s)", serviceName, err, strings.TrimSpace(out))
	}
	if plistRunningPID(out) {
		return StatusRunning, nil
	}
	return StatusStopped, nil
}
