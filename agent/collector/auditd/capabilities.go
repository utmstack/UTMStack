//go:build linux
// +build linux

package auditd

import (
	"os/exec"
	"strings"

	"github.com/utmstack/UTMStack/agent/utils"
)

// checkAuditCapability checks if the audit system is available and enabled.
// Uses auditctl -s to verify audit status since /proc/sys/kernel/auditing
// doesn't exist on all kernel versions.
func checkAuditCapability() error {
	// Check if auditctl exists
	auditctlPath, err := exec.LookPath("auditctl")
	if err != nil {
		utils.Logger.ErrorF("auditd: auditctl not found in PATH: %v", err)
		return err
	}

	// Run auditctl -s to check audit status
	cmd := exec.Command(auditctlPath, "-s")
	output, err := cmd.Output()
	if err != nil {
		utils.Logger.ErrorF("auditd: failed to run auditctl -s: %v", err)
		return err
	}

	// Check if enabled=1 in output
	if !strings.Contains(string(output), "enabled 1") && !strings.Contains(string(output), "enabled=1") {
		utils.Logger.Info("auditd: kernel auditing is disabled (enabled != 1), collector will not start")
		return nil
	}

	utils.Logger.Info("auditd: audit system is enabled and ready")
	return nil
}
