package config

import (
	"path/filepath"
	"runtime"

	"github.com/utmstack/UTMStack/agent/installer/utils"
)

const (
	DEPEND_URL         = "https://%s/dependencies/agent?version=%s&os=%s&type=%s"
	INSTALLER_LOG_FILE = "utmstack_agent_installer.log"
	UpdaterSelfLinux   = "utmstack_updater_self"
	AgentManagerPort   = "9000"
	LogAuthProxyPort   = "50051"
	DependServiceLabel = "service"
	DependZipLabel     = "depend_zip"
)

func GetCertPath() string {
	path := utils.GetMyPath()
	return filepath.Join(path, "certs", "utm.crt")
}

func GetServiceBin() string {
	switch runtime.GOOS {
	case "windows":
		return "utmstack_agent_service.exe"
	case "linux":
		return "utmstack_agent_service"
	}
	return ""
}

func GetDownloadFilePath(typ string) string {
	path := utils.GetMyPath()
	switch typ {
	case DependServiceLabel:
		switch runtime.GOOS {
		case "windows":
			return filepath.Join(path, "utmstack_agent_service.exe")
		case "linux":
			return filepath.Join(path, "utmstack_agent_service")
		}
	case DependZipLabel:
		return filepath.Join(path, "dependencies.zip")
	}
	return ""
}

func GetVersionPath() string {
	path := utils.GetMyPath()
	return filepath.Join(path, "version.json")
}
