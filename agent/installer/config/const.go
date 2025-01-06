package config

import (
	"path/filepath"
	"runtime"

	"github.com/utmstack/UTMStack/agent/installer/utils"
)

var (
	DependUrl        = "https://%s:%s/private/dependencies/%s"
	VersionUrl       = "https://%s:%s/private/version?service=%s"
	InstallerLogFile = "utmstack_agent_installer.log"
	UpdaterSelfLinux = "utmstack_updater_self"
	AgentManagerPort = "9000"
	DependenciesPort = "9001"
	LogAuthProxyPort = "50051"
	VersionPath      = filepath.Join(utils.GetMyPath(), "version.json")
)

func GetDependFiles() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{
			"utmstack_agent_service.exe",
			"utmstack_agent_dependencies_windows.zip",
		}
	case "linux":
		return []string{
			"utmstack_agent_service",
			"utmstack_agent_dependencies_linux.zip",
		}
	}
	return nil
}
