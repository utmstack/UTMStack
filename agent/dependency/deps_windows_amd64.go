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
		// TODO: Remove beats dependency after testing native collectors
		// {
		// 	Name:         "beats",
		// 	Version:      BeatsVersion,
		// 	BinaryPath:   filepath.Join(basePath, "beats", "filebeat", "filebeat.exe"),
		// 	DownloadName: beatsZip,
		// 	DownloadURL: func(server string) string {
		// 		return fmt.Sprintf(config.DependUrl, server, config.DependenciesPort, beatsZip)
		// 	},
		// 	Critical: false,
		// 	PostDownload: func() error {
		// 		zipPath := filepath.Join(basePath, beatsZip)
		// 		if err := archive.Unzip(zipPath, basePath); err != nil {
		// 			return fmt.Errorf("error unzipping beats: %v", err)
		// 		}
		// 		// Remove zip after extraction
		// 		os.Remove(zipPath)
		// 		return nil
		// 	},
		// 	Configure: configureBeats,
		// 	Uninstall: uninstallBeats,
		// },

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

// TODO: Remove after testing native collectors
// func configureBeats() error {
// 	return collector.InstallAll()
// }

func uninstallBeats() error {
	return collector.UninstallAll()
}
