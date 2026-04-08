//go:build linux && amd64
// +build linux,amd64

package dependency

import (
	"fmt"
	"path/filepath"

	"github.com/utmstack/UTMStack/agent/collector"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/shared/exec"
	"github.com/utmstack/UTMStack/shared/fs"
)

// GetDependencies returns the list of dependencies for Linux amd64.
func GetDependencies() []Dependency {
	basePath := fs.GetExecutablePath()

	return []Dependency{
		{
			Name:       "updater",
			Version:    UpdaterVersion,
			BinaryPath: filepath.Join(basePath, UpdaterFile("")),
			DownloadURL: func(server string) string {
				return fmt.Sprintf(config.DependUrl, server, config.DependenciesPort, UpdaterFile(""))
			},
			Critical:  false,
			Configure: configureUpdater,
			Uninstall: uninstallUpdater,
		},

		// New beats dependency - only for uninstalling existing filebeat/winlogbeat
		// No download, no install - native collectors are used instead
		{
			Name:      "beats",
			Version:   BeatsVersion,
			Critical:  false,
			Uninstall: uninstallBeats,
		},

		// Auditd dependency - auto-configures Linux audit daemon
		// No download - installs from system package manager
		{
			Name:       "auditd",
			Version:    AuditdVersion,
			BinaryPath: "/sbin/auditctl", // Check if auditd tools exist
			Critical:   false,
			Configure:  configureAuditd,
			Update:     updateAuditdRules,
			Uninstall:  cleanupAuditd,
		},
	}
}

func configureUpdater() error {
	updaterPath := filepath.Join(fs.GetExecutablePath(), UpdaterFile(""))

	// Set executable permissions
	if err := exec.Run("chmod", fs.GetExecutablePath(), "755", updaterPath); err != nil {
		return fmt.Errorf("error setting permissions on updater: %v", err)
	}

	return exec.Run(updaterPath, fs.GetExecutablePath(), "install")
}

func uninstallUpdater() error {
	updaterPath := filepath.Join(fs.GetExecutablePath(), UpdaterFile(""))
	if !fs.Exists(updaterPath) {
		return nil
	}
	return exec.Run(updaterPath, fs.GetExecutablePath(), "uninstall")
}

func uninstallBeats() error {
	return collector.UninstallAll()
}
