//go:build windows
// +build windows

package svc

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const (
	pollInterval = 500 * time.Millisecond
	stopTimeout  = 60 * time.Second
	startTimeout = 60 * time.Second
)

// Start starts a Windows service by name and waits until it's running.
func Start(serviceName string) error {
	// Already running? Nothing to do
	if running, _ := IsActive(serviceName); running {
		return nil
	}

	cmd := exec.Command("sc", "start", serviceName)
	cmd.Run() // Ignore error, we'll check actual state

	// Poll until running or timeout
	deadline := time.Now().Add(startTimeout)
	for time.Now().Before(deadline) {
		if running, _ := IsActive(serviceName); running {
			return nil
		}
		time.Sleep(pollInterval)
	}

	return fmt.Errorf("timeout waiting for service %s to start", serviceName)
}

// Stop stops a Windows service by name and waits until it's stopped.
func Stop(serviceName string) error {
	// Already stopped? Nothing to do
	status, _ := Status(serviceName)
	if status == StatusStopped {
		return nil
	}

	cmd := exec.Command("sc", "stop", serviceName)
	cmd.Run() // Ignore error, we'll check actual state

	// Poll until stopped or timeout
	deadline := time.Now().Add(stopTimeout)
	for time.Now().Before(deadline) {
		status, _ := Status(serviceName)
		if status == StatusStopped {
			return nil
		}
		time.Sleep(pollInterval)
	}

	return fmt.Errorf("timeout waiting for service %s to stop", serviceName)
}

// Restart restarts a Windows service by stopping and starting it.
func Restart(serviceName string) error {
	if err := Stop(serviceName); err != nil {
		return err
	}
	return Start(serviceName)
}

// IsActive checks if a Windows service is running.
func IsActive(serviceName string) (bool, error) {
	status, err := Status(serviceName)
	if err != nil {
		return false, err
	}
	return status == StatusRunning, nil
}

// Status returns the status of a Windows service.
func Status(serviceName string) (string, error) {
	cmd := exec.Command("sc", "query", serviceName)
	output, err := cmd.Output()
	if err != nil {
		return StatusUnknown, fmt.Errorf("failed to query service %s: %w", serviceName, err)
	}

	outputStr := string(output)
	switch {
	case strings.Contains(outputStr, "RUNNING"):
		return StatusRunning, nil
	case strings.Contains(outputStr, "STOPPED"):
		return StatusStopped, nil
	default:
		return StatusUnknown, nil
	}
}

// EnsureRecovery configures Windows Service Control Manager failure actions so
// the service automatically restarts after unexpected crashes or non-zero exits.
func EnsureRecovery(serviceName string) error {
	if out, err := exec.Command("sc", "failure", serviceName, "reset=", "86400", "actions=", "restart/5000/restart/15000/restart/60000").CombinedOutput(); err != nil {
		return fmt.Errorf("failed to configure failure recovery for %s: %w (%s)", serviceName, err, strings.TrimSpace(string(out)))
	}
	if out, err := exec.Command("sc", "failureflag", serviceName, "1").CombinedOutput(); err != nil {
		return fmt.Errorf("failed to configure failure flag for %s: %w (%s)", serviceName, err, strings.TrimSpace(string(out)))
	}
	return nil
}
