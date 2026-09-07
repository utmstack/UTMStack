//go:build windows && amd64
// +build windows,amd64

package dependency

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/utils"
	"github.com/utmstack/UTMStack/shared/exec"
	"github.com/utmstack/UTMStack/shared/fs"
	"github.com/utmstack/UTMStack/shared/svc"
)

// GetDependencies returns the list of dependencies for Windows amd64.
func GetDependencies() []Dependency {
	basePath := fs.GetExecutablePath()

	return []Dependency{
		{
			Name:        "updater",
			Version:     getUpdaterVersion(),
			BinaryPath:  filepath.Join(basePath, UpdaterFile("")),
			DownloadURL: func(server string) string {
				return fmt.Sprintf(config.DependUrl, server, config.DependenciesPort, UpdaterFile(""))
			},
			Critical:    false, // Agent can run without updater
			PreDownload: preDownloadUpdater,
			Configure:   configureUpdater,
			Uninstall:   uninstallUpdater,
		},

		// Windows Advanced Audit Policy - ensures the Security/PowerShell
		// channels this agent collects actually get populated (see A9 in
		// agent/GAPS_AND_IMPROVEMENTS.md). No download - configures the
		// OS's own audit policy and a few registry settings.
		{
			Name:       "audit-policy",
			Version:    AuditPolicyVersion,
			BinaryPath: filepath.Join(os.Getenv("windir"), "System32", "auditpol.exe"),
			Critical:   false,
			Configure:  configureWindowsAuditPolicy,
			Update:     configureWindowsAuditPolicy,
		},
	}
}

func configureUpdater() error {
	updaterPath := filepath.Join(fs.GetExecutablePath(), UpdaterFile(""))
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
	_ = utils.StopService(config.ModulesServName)
	_ = utils.UninstallService(config.ModulesServName)
	_ = utils.StopService("UTMStackWindowsLogsCollector")
	_ = utils.UninstallService("UTMStackWindowsLogsCollector")
	return nil
}

func preDownloadUpdater() (func(), error) {
	// Stop the updater service before download
	if err := svc.Stop(config.SERVICE_UPDATER_NAME); err != nil {
		// Service might not be running or installed yet - that's OK
		// Return cleanup function anyway (safe to start)
		return func() {
			_ = svc.Start(config.SERVICE_UPDATER_NAME)
		}, nil
	}
	
	// Return cleanup function that restarts the service
	return func() {
		_ = svc.Start(config.SERVICE_UPDATER_NAME)
	}, nil
}
