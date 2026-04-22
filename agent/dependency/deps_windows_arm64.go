//go:build windows && arm64
// +build windows,arm64

package dependency

import (
	"fmt"
	"path/filepath"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/shared/exec"
	"github.com/utmstack/UTMStack/shared/fs"
	"github.com/utmstack/UTMStack/shared/svc"
)

// GetDependencies returns the list of dependencies for Windows arm64.
// Windows arm64 uses integrated Windows collector, no beats needed.
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
			Critical:    false,
			PreDownload: preDownloadUpdater,
			Configure:   configureUpdater,
			Uninstall:   uninstallUpdater,
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
