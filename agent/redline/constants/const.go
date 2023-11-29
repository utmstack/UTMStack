package constants

import "runtime"

const (
	SERV_NAME      = "UTMStackRedline"
	SERV_LOG       = "utmstack_redline.log"
	SERV_LOCK_NAME = "utmstack_redline.lock"
)

func GetServicesLock() map[string]string {
	var services = map[string]string{
		"UTMStackAgent":                "utmstack_agent.lock",
		"UTMStackModulesLogsCollector": "utmstack_modules_collector.lock",
		"UTMStackUpdater":              "utmstack_updater.lock",
	}
	if runtime.GOOS == "windows" {
		services["UTMStackWindowsLogsCollector"] = "utmstack_windows_collector.lock"
	}
	return services
}

func GetAgentBin() string {
	var bin string
	switch runtime.GOOS {
	case "windows":
		bin = "utmstack_agent_service.exe"
	case "linux":
		bin = "utmstack_agent_service"
	}
	return bin
}
