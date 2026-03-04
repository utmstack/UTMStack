//go:build windows
// +build windows

package svc

import (
	"fmt"
	"os/exec"
	"strings"
)

// Start starts a Windows service by name.
func Start(serviceName string) error {
	cmd := exec.Command("sc", "start", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start service %s: %w", serviceName, err)
	}
	return nil
}

// Stop stops a Windows service by name.
func Stop(serviceName string) error {
	cmd := exec.Command("sc", "stop", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop service %s: %w", serviceName, err)
	}
	return nil
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
	cmd := exec.Command("sc", "query", serviceName)
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to query service %s: %w", serviceName, err)
	}
	return strings.Contains(string(output), "RUNNING"), nil
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
