//go:build linux || darwin
// +build linux darwin

package svc

import (
	"fmt"
	"os/exec"
	"strings"
)

// Start starts a system service by name.
func Start(serviceName string) error {
	cmd := exec.Command("systemctl", "start", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start service %s: %w", serviceName, err)
	}
	return nil
}

// Stop stops a system service by name.
func Stop(serviceName string) error {
	cmd := exec.Command("systemctl", "stop", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop service %s: %w", serviceName, err)
	}
	return nil
}

// Restart restarts a system service by name.
func Restart(serviceName string) error {
	cmd := exec.Command("systemctl", "restart", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restart service %s: %w", serviceName, err)
	}
	return nil
}

// IsActive checks if a system service is running.
func IsActive(serviceName string) (bool, error) {
	cmd := exec.Command("systemctl", "is-active", serviceName)
	output, err := cmd.Output()
	if err != nil {
		// is-active returns non-zero for inactive services
		return false, nil
	}
	return strings.TrimSpace(string(output)) == "active", nil
}

// Status returns the status of a system service.
func Status(serviceName string) (string, error) {
	cmd := exec.Command("systemctl", "is-active", serviceName)
	output, _ := cmd.Output()
	status := strings.TrimSpace(string(output))

	switch status {
	case "active":
		return StatusRunning, nil
	case "inactive", "failed":
		return StatusStopped, nil
	default:
		return StatusUnknown, nil
	}
}
