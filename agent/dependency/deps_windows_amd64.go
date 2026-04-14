//go:build windows && amd64
// +build windows,amd64

package dependency

import (
	"fmt"
	"path/filepath"

	"github.com/utmstack/UTMStack/agent/collector"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/shared/exec"
	"github.com/utmstack/UTMStack/shared/fs"
)

// GetDependencies returns the list of dependencies for Windows amd64.
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
			Critical:  false, // Agent can run without updater
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
	return collector.UninstallAll()
}
