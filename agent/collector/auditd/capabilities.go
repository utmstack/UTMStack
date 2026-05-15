//go:build linux
// +build linux

package auditd

import (
	"errors"
	"os/exec"
	"strings"

	"github.com/utmstack/UTMStack/agent/utils"
)

var ErrAuditUnavailable = errors.New("audit subsystem unavailable in this environment")

// checkAuditCapability checks if the audit system is available and enabled.
// Uses auditctl -s to verify audit status since /proc/sys/kernel/auditing
// doesn't exist on all kernel versions.
func checkAuditCapability() error {
	auditctlPath, err := exec.LookPath("auditctl")
	if err != nil {
		utils.Logger.Info("auditd: auditctl not found in PATH, collector will not start")
		return ErrAuditUnavailable
	}

	cmd := exec.Command(auditctlPath, "-s")
	output, err := cmd.Output()
	if err != nil {
		utils.Logger.Info("auditd: failed to run auditctl -s (%v), collector will not start", err)
		return ErrAuditUnavailable
	}

	if !strings.Contains(string(output), "enabled 1") && !strings.Contains(string(output), "enabled=1") {
		utils.Logger.Info("auditd: kernel auditing is disabled (enabled != 1), collector will not start")
		return ErrAuditUnavailable
	}

	utils.Logger.Info("auditd: audit system is enabled and ready")
	return nil
}
