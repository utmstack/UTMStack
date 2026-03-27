//go:build linux
// +build linux

package dependency

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/utmstack/UTMStack/agent/utils"
	sharedExec "github.com/utmstack/UTMStack/shared/exec"
	"github.com/utmstack/UTMStack/shared/fs"
)

// AuditdVersion is the UTMStack audit rules version.
// Bump this when audit rules are updated to trigger rule redeployment.
const AuditdVersion = "1.0.0"

// auditRulesContent contains the UTMStack security audit rules.
// These rules are deployed to /etc/audit/rules.d/50-utmstack.rules
// NOTE: This is an ADDITIVE configuration - does not delete existing rules
// or modify global settings (-D, -b, -f) to coexist with enterprise policies.
const auditRulesContent = `## UTMStack SIEM Audit Rules
## Managed by UTMStack Agent - Do not edit manually
## Additive rules - does not delete existing configuration

# Monitor executed commands (critical for SIEM)
-a always,exit -F arch=b64 -S execve -k utmstack_exec
-a always,exit -F arch=b32 -S execve -k utmstack_exec

# Privilege escalation
-a always,exit -F arch=b64 -S setuid,setgid,setreuid,setregid,setresuid,setresgid -F auid>=1000 -k utmstack_priv
-a always,exit -F arch=b32 -S setuid,setgid,setreuid,setregid,setresuid,setresgid -F auid>=1000 -k utmstack_priv

# Sensitive file access
-w /etc/shadow -p wa -k utmstack_sensitive
-w /etc/passwd -p wa -k utmstack_sensitive
-w /etc/sudoers -p wa -k utmstack_sensitive
-w /etc/sudoers.d -p wa -k utmstack_sensitive
-w /etc/ssh/sshd_config -p wa -k utmstack_sensitive
-w /root/.ssh -p wa -k utmstack_sensitive

# Module loading
-a always,exit -F arch=b64 -S init_module,finit_module,delete_module -k utmstack_modules
-a always,exit -F arch=b32 -S init_module,finit_module,delete_module -k utmstack_modules

# Network connections (may be high volume - consider enabling selectively)
# -a always,exit -F arch=b64 -S connect -F auid>=1000 -k utmstack_network

# Time changes
-a always,exit -F arch=b64 -S adjtimex,settimeofday,clock_settime -k utmstack_time
-a always,exit -F arch=b32 -S adjtimex,settimeofday,clock_settime -k utmstack_time

# Audit configuration changes
-w /etc/audit -p wa -k utmstack_audit_config
-w /etc/audisp -p wa -k utmstack_audit_config
`

const (
	auditRulesPath = "/etc/audit/rules.d/50-utmstack.rules"
)

// configureAuditd is called by dependency.Reconcile on first install.
// It orchestrates the full auditd setup: root check → container check → distro detect → install → rules → service
func configureAuditd() error {
	// Check if running as root
	if !isRoot() {
		utils.Logger.Info("auditd setup skipped: root privileges required")
		return nil
	}

	// Check if running in a container
	if isContainer() {
		utils.Logger.Info("auditd setup skipped: container environment detected")
		return nil
	}

	// Detect distribution
	distro := DetectDistro()
	if distro.PackageManager == "" {
		utils.Logger.Info("auditd setup skipped: unknown distribution, no package manager detected")
		return nil
	}

	utils.Logger.Info("Detected distro: ID=%s, IDLike=%s, PackageManager=%s", distro.ID, distro.IDLike, distro.PackageManager)

	// Install auditd if not already installed
	if !isAuditdInstalled() {
		utils.Logger.Info("Installing auditd package...")
		if err := installAuditd(distro); err != nil {
			utils.Logger.ErrorF("Failed to install auditd: %v", err)
			return nil // Non-critical, don't fail the agent
		}
		utils.Logger.Info("auditd package installed successfully")
	} else {
		utils.Logger.Info("auditd is already installed")
	}

	// Pre-flight check: can we modify audit configuration?
	if canConfigure, reason := canConfigureAuditd(); !canConfigure {
		utils.Logger.Info("auditd rule deployment skipped: %s", reason)
		return nil // Non-critical, don't fail the agent
	}

	// Deploy audit rules
	utils.Logger.Info("Deploying UTMStack audit rules...")
	if err := deployRules(); err != nil {
		utils.Logger.ErrorF("Failed to deploy audit rules: %v", err)
		return nil // Non-critical, don't fail the agent
	}
	utils.Logger.Info("UTMStack audit rules deployed successfully")

	// Start and enable auditd service
	utils.Logger.Info("Starting auditd service...")
	if err := startAuditd(); err != nil {
		utils.Logger.ErrorF("Failed to start auditd service: %v", err)
		return nil // Non-critical, don't fail the agent
	}
	utils.Logger.Info("auditd service started and enabled")

	// Reload rules
	utils.Logger.Info("Reloading audit rules...")
	if err := reloadRules(); err != nil {
		utils.Logger.ErrorF("Failed to reload audit rules: %v", err)
		return nil // Non-critical, don't fail the agent
	}
	utils.Logger.Info("Audit rules reloaded successfully")

	return nil
}

// updateAuditdRules is called when AuditdVersion changes.
// It redeploys rules without reinstalling the package.
func updateAuditdRules() error {
	// Check if running as root
	if !isRoot() {
		utils.Logger.Info("auditd rule update skipped: root privileges required")
		return nil
	}

	// Check if running in a container
	if isContainer() {
		utils.Logger.Info("auditd rule update skipped: container environment detected")
		return nil
	}

	// Pre-flight check: can we modify audit configuration?
	if canConfigure, reason := canConfigureAuditd(); !canConfigure {
		utils.Logger.Info("auditd rule update skipped: %s", reason)
		return nil // Non-critical, don't fail the agent
	}

	// Deploy updated rules
	utils.Logger.Info("Updating UTMStack audit rules...")
	if err := deployRules(); err != nil {
		utils.Logger.ErrorF("Failed to deploy audit rules: %v", err)
		return nil
	}

	// Reload rules
	if err := reloadRules(); err != nil {
		utils.Logger.ErrorF("Failed to reload audit rules: %v", err)
		return nil
	}

	utils.Logger.Info("UTMStack audit rules updated successfully")
	return nil
}

// isRoot checks if the current process is running as root.
func isRoot() bool {
	return os.Geteuid() == 0
}

// isContainer checks if the process is running inside a container.
// Checks for Docker, Podman, LXC, and other container runtimes.
func isContainer() bool {
	// Check for Docker/Podman environment files
	containerFiles := []string{
		"/.dockerenv",
		"/run/.containerenv",
	}
	for _, f := range containerFiles {
		if fs.Exists(f) {
			return true
		}
	}

	// Check cgroup for container patterns
	cgroupPath := "/proc/1/cgroup"
	if content, err := os.ReadFile(cgroupPath); err == nil {
		cgroupStr := string(content)
		containerPatterns := []string{"docker", "kubepods", "lxc", "containerd"}
		for _, pattern := range containerPatterns {
			if strings.Contains(cgroupStr, pattern) {
				return true
			}
		}
	}

	return false
}

// canConfigureAuditd performs pre-flight checks to determine if audit rules can be modified.
// Returns (true, "") if configuration is possible, (false, "reason") if not.
// This checks for:
// - Immutable mode (-e 2): audit config is locked until reboot
// - never,task rule: auditing is disabled system-wide
func canConfigureAuditd() (bool, string) {
	// Check if auditctl is available
	auditctlPath, err := exec.LookPath("auditctl")
	if err != nil {
		// auditctl not found - this is OK, we'll install auditd first
		return true, ""
	}

	// Check for immutable mode (-e 2)
	// auditctl -s returns status including "enabled X" where X is 0, 1, or 2
	statusCmd := exec.Command(auditctlPath, "-s")
	statusOutput, err := statusCmd.Output()
	if err == nil {
		statusStr := string(statusOutput)
		if strings.Contains(statusStr, "enabled 2") {
			return false, "audit config is locked (immutable mode -e 2), reboot required to modify"
		}
	}

	// Check for never,task rule (auditing disabled)
	// auditctl -l lists all loaded rules
	rulesCmd := exec.Command(auditctlPath, "-l")
	rulesOutput, err := rulesCmd.Output()
	if err == nil {
		rulesStr := string(rulesOutput)
		if strings.Contains(rulesStr, "never,task") {
			return false, "auditing is disabled by never,task rule"
		}
	}

	return true, ""
}

// isAuditdInstalled checks if auditd tools are installed.
func isAuditdInstalled() bool {
	// Check common paths for auditctl
	paths := []string{
		"/sbin/auditctl",
		"/usr/sbin/auditctl",
		"/usr/bin/auditctl",
	}
	for _, p := range paths {
		if fs.Exists(p) {
			return true
		}
	}

	// Also try exec.LookPath as fallback
	if _, err := exec.LookPath("auditctl"); err == nil {
		return true
	}

	return false
}

// isAuditdRunning checks if the auditd service is currently running.
func isAuditdRunning() bool {
	// Try systemctl first
	if _, err := exec.LookPath("systemctl"); err == nil {
		cmd := exec.Command("systemctl", "is-active", "--quiet", "auditd")
		if err := cmd.Run(); err == nil {
			return true
		}
	}

	// Fallback: check if auditd process exists
	if content, err := os.ReadFile("/proc/1/comm"); err == nil {
		if strings.TrimSpace(string(content)) == "auditd" {
			return true
		}
	}

	// Check /var/run/auditd.pid
	if fs.Exists("/var/run/auditd.pid") {
		return true
	}

	return false
}

// installAuditd installs the auditd package using the appropriate package manager.
func installAuditd(distro *DistroInfo) error {
	var cmd *exec.Cmd

	switch distro.PackageManager {
	case "apt":
		// Debian/Ubuntu: auditd + audispd-plugins
		cmd = exec.Command("apt-get", "install", "-y", "auditd", "audispd-plugins")
		// Set DEBIAN_FRONTEND to avoid prompts
		cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")

	case "dnf":
		// Fedora, RHEL 8+, Rocky, Alma
		cmd = exec.Command("dnf", "install", "-y", "audit")

	case "yum":
		// RHEL 7, CentOS 7, Amazon Linux
		cmd = exec.Command("yum", "install", "-y", "audit")

	case "zypper":
		// SUSE, openSUSE
		cmd = exec.Command("zypper", "install", "-y", "audit")

	case "pacman":
		// Arch, Manjaro
		cmd = exec.Command("pacman", "-S", "--noconfirm", "audit")

	default:
		return fmt.Errorf("unsupported package manager: %s", distro.PackageManager)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("package installation failed: %v, output: %s", err, string(output))
	}

	return nil
}

// startAuditd enables and starts the auditd service.
func startAuditd() error {
	// Try systemctl first (most common on modern systems)
	if _, err := exec.LookPath("systemctl"); err == nil {
		// Enable the service
		if err := sharedExec.Run("systemctl", "/", "enable", "auditd"); err != nil {
			utils.Logger.LogF(400, "Failed to enable auditd service: %v", err)
			// Continue anyway, try to start
		}

		// Start the service
		if err := sharedExec.Run("systemctl", "/", "start", "auditd"); err != nil {
			// auditd might already be running, which is fine
			if !isAuditdRunning() {
				return fmt.Errorf("failed to start auditd service: %v", err)
			}
		}
		return nil
	}

	// Fallback to service command
	if _, err := exec.LookPath("service"); err == nil {
		if err := sharedExec.Run("service", "/", "auditd", "start"); err != nil {
			if !isAuditdRunning() {
				return fmt.Errorf("failed to start auditd service: %v", err)
			}
		}
		return nil
	}

	return fmt.Errorf("no service manager found (systemctl or service)")
}

// deployRules writes the UTMStack audit rules to the rules.d directory.
func deployRules() error {
	// Ensure the rules.d directory exists
	rulesDir := filepath.Dir(auditRulesPath)
	if err := fs.CreateDirIfNotExist(rulesDir); err != nil {
		return fmt.Errorf("failed to create rules directory: %v", err)
	}

	// Write the rules file
	if err := fs.WriteString(auditRulesPath, auditRulesContent); err != nil {
		return fmt.Errorf("failed to write rules file: %v", err)
	}

	return nil
}

// cleanupAuditd removes UTMStack audit rules on agent uninstall.
// It removes the rules file and reloads auditd to clear the rules from kernel.
// Does NOT stop auditd service or uninstall the package.
func cleanupAuditd() error {
	// Skip if not root
	if !isRoot() {
		utils.Logger.Info("skipping auditd cleanup: not running as root")
		return nil
	}

	// Skip if running in container
	if isContainer() {
		utils.Logger.Info("skipping auditd cleanup: running in container")
		return nil
	}

	// Remove our rules file if it exists
	if _, err := os.Stat(auditRulesPath); err == nil {
		if err := os.Remove(auditRulesPath); err != nil {
			utils.Logger.ErrorF("failed to remove auditd rules file: %v", err)
			return err
		}
		utils.Logger.Info("removed UTMStack auditd rules file")

		// Reload auditd rules to clear our rules from kernel
		if err := reloadRules(); err != nil {
			utils.Logger.ErrorF("failed to reload auditd rules after cleanup: %v", err)
			// Don't fail - the file is already removed
		}
	}

	// Do NOT stop auditd service
	// Do NOT uninstall auditd package

	return nil
}

// reloadRules loads the audit rules into the kernel.
func reloadRules() error {
	// Try augenrules first (preferred method for rules.d)
	if _, err := exec.LookPath("augenrules"); err == nil {
		if err := sharedExec.Run("augenrules", "/", "--load"); err != nil {
			utils.Logger.LogF(400, "augenrules --load failed: %v, trying auditctl", err)
		} else {
			return nil
		}
	}

	// Fallback to auditctl -R
	if _, err := exec.LookPath("auditctl"); err == nil {
		if err := sharedExec.Run("auditctl", "/", "-R", auditRulesPath); err != nil {
			return fmt.Errorf("failed to load rules with auditctl: %v", err)
		}
		return nil
	}

	return fmt.Errorf("no audit rule loader found (augenrules or auditctl)")
}
