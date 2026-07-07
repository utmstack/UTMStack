package dependency

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/shared/exec"
	"github.com/utmstack/UTMStack/shared/fs"
)

const EDRVersion = "12.0.0"

// EDRFile returns the EDR binary name with OS/arch suffix, matching UpdaterFile.
func EDRFile(suffix string) string {
	name := fmt.Sprintf("utmstack_edr_%s_%s%s", runtime.GOOS, runtime.GOARCH, suffix)
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func configureEDR() error {
	edrPath := filepath.Join(fs.GetExecutablePath(), EDRFile(""))
	return exec.Run(edrPath, fs.GetExecutablePath(), "install")
}

func uninstallEDR() error {
	edrPath := filepath.Join(fs.GetExecutablePath(), EDRFile(""))
	if !fs.Exists(edrPath) {
		return nil
	}
	return exec.Run(edrPath, fs.GetExecutablePath(), "uninstall")
}

func edrDownloadURL(server string) string {
	return fmt.Sprintf(config.DependUrl, server, config.DependenciesPort, EDRFile(""))
}
