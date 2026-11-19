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

// darwinStopRetry lets launchd finish tearing the daemon down before Restart
// reloads it.
const darwinStopRetry = 50 * time.Millisecond

// runDarwin executes a command and returns its combined output; the error is
// non-nil iff the process exits non-zero (exec.Cmd.CombinedOutput semantics).
// Overridden in tests.
var runDarwin = func(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	return string(out), err
}

// plistInstalled reports whether the service's launch daemon plist exists.
// Overridden in tests.
var plistInstalled = func(serviceName string) bool {
	_, err := os.Stat(daemonPlistPath(serviceName))
	return err == nil
}

// daemonLoaded reports whether the service is registered in launchd.
// `launchctl list <label>` exits 0 when loaded and non-zero (113) when not,
// so the exit code is the deterministic signal regardless of the message text
// (which varies by macOS version).
func daemonLoaded(serviceName string) bool {
	_, err := runDarwin("launchctl", "list", serviceName)
	return err == nil
}

// Start starts a launch daemon by loading its plist.
func Start(serviceName string) error {
	plistPath := daemonPlistPath(serviceName)
	if !plistInstalled(serviceName) {
		return fmt.Errorf("service %s is not installed: %s", serviceName, plistPath)
	}
	if out, err := runDarwin("launchctl", "load", plistPath); err != nil {
		return fmt.Errorf("failed to start service %s: %v (%s)", serviceName, err, strings.TrimSpace(out))
	}
	return nil
}

// Stop stops a launch daemon by unloading its plist. Unloading a daemon that
// is not currently loaded fails; that is a successful no-op (the updater
// stops the agent before swapping the binary, which may already be down), so
// the error is ignored only when the daemon is confirmed unloaded afterwards.
func Stop(serviceName string) error {
	plistPath := daemonPlistPath(serviceName)
	if !plistInstalled(serviceName) {
		return fmt.Errorf("service %s is not installed: %s", serviceName, plistPath)
	}
	out, err := runDarwin("launchctl", "unload", plistPath)
	if err == nil {
		return nil
	}
	if !daemonLoaded(serviceName) {
		return nil
	}
	return fmt.Errorf("failed to stop service %s: %v (%s)", serviceName, err, strings.TrimSpace(out))
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
// is not running, without error (the updater health-checks right after Start).
func IsActive(serviceName string) (bool, error) {
	out, err := runDarwin("launchctl", "list", serviceName)
	if err != nil {
		return false, nil
	}
	return plistRunningPID(out), nil
}

// Status returns the status of a launch daemon.
func Status(serviceName string) (string, error) {
	out, err := runDarwin("launchctl", "list", serviceName)
	if err != nil {
		// Not loaded: stopped if installed, unknown otherwise.
		if !plistInstalled(serviceName) {
			return StatusUnknown, nil
		}
		return StatusStopped, nil
	}
	if plistRunningPID(out) {
		return StatusRunning, nil
	}
	return StatusStopped, nil
}
