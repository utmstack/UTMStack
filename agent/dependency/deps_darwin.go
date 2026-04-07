//go:build darwin
// +build darwin

package dependency

import (
	"fmt"
	"path/filepath"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/shared/exec"
	"github.com/utmstack/UTMStack/shared/fs"
)

const macosCollectorBinary = "utmstack-collector-mac"

// GetDependencies returns the list of dependencies for macOS.
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
		{
			Name:       "macos-collector",
			Version:    MacosCollectorVersion,
			BinaryPath: filepath.Join(basePath, macosCollectorBinary),
			DownloadURL: func(server string) string {
				return fmt.Sprintf(config.DependUrl, server, config.DependenciesPort, macosCollectorBinary)
			},
			Critical:  true,
			Configure: configureMacosCollector,
		},
	}
}

func configureMacosCollector() error {
	collectorPath := filepath.Join(fs.GetExecutablePath(), macosCollectorBinary)
	return exec.Run("chmod", fs.GetExecutablePath(), "755", collectorPath)
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
