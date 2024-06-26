package configuration

import (
	"path/filepath"
	"runtime"

	"github.com/utmstack/UTMStack/agent/installer/utils"
)

const (
	INSTALLER_LOG_FILE = "utmstack_agent_installer.log"
	AgentManagerPort   = "9000"
	LogAuthProxyPort   = "50051"
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
	case "service":
		switch runtime.GOOS {
		case "windows":
			return filepath.Join(path, "utmstack_agent_service.exe")
		case "linux":
			return filepath.Join(path, "utmstack_agent_service")
		}
	case "dependencies":
		return filepath.Join(path, "dependencies.zip")
	}
	return ""
}

func GetVersionPath() string {
	path := utils.GetMyPath()
	return filepath.Join(path, "version.json")
}
