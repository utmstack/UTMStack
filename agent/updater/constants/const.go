package constants

import (
	"path/filepath"
	"runtime"
	"time"

	"github.com/utmstack/UTMStack/agent/updater/utils"
)

const (
	SERV_NAME             = "UTMStackUpdater"
	SERV_LOG              = "utmstack_updater.log"
	SERV_LOCK             = "utmstack_updater.lock"
	MASTERVERSIONENDPOINT = "/management/info"
	Bucket                = "https://cdn.utmstack.com/agent_updates/"
)

type ServiceAttribt struct {
	ServName string
	ServBin  string
	ServLock string
}

var (
	CheckUpdatesEvery = time.Duration(5 * time.Minute)
)

func GetServAttr() map[string]ServiceAttribt {
	serAttr := map[string]ServiceAttribt{}

	switch runtime.GOOS {
	case "windows":
		serAttr["agent"] = ServiceAttribt{ServName: "UTMStackAgent", ServBin: "utmstack_agent_service.exe", ServLock: "utmstack_agent.lock"}
		serAttr["filebeat"] = ServiceAttribt{ServName: "UTMStackModulesLogsCollector", ServBin: "utmstack_agent_service.exe", ServLock: "utmstack_modules_collector.lock"}
		serAttr["winlogbeat"] = ServiceAttribt{ServName: "UTMStackWindowsLogsCollector", ServBin: "utmstack_agent_service.exe", ServLock: "utmstack_windows_collector.lock"}
		serAttr["updater"] = ServiceAttribt{ServName: "UTMStackUpdater", ServBin: "utmstack_updater_self.exe", ServLock: "utmstack_updater.lock"}
		serAttr["redline"] = ServiceAttribt{ServName: "UTMStackRedline", ServBin: "utmstack_redline_service.exe", ServLock: "utmstack_redline.lock"}
	case "linux":
		serAttr["agent"] = ServiceAttribt{ServName: "UTMStackAgent", ServBin: "utmstack_agent_service", ServLock: "utmstack_agent.lock"}
		serAttr["filebeat"] = ServiceAttribt{ServName: "UTMStackModulesLogsCollector", ServBin: "utmstack_agent_service", ServLock: "utmstack_modules_collector.lock"}
		serAttr["winlogbeat"] = ServiceAttribt{ServName: "UTMStackWindowsLogsCollector", ServBin: "utmstack_agent_service", ServLock: "utmstack_windows_collector.lock"}
		serAttr["updater"] = ServiceAttribt{ServName: "UTMStackUpdater", ServBin: "utmstack_updater_self", ServLock: "utmstack_updater.lock"}
		serAttr["redline"] = ServiceAttribt{ServName: "UTMStackRedline", ServBin: "utmstack_redline_service", ServLock: "utmstack_redline.lock"}
	}
	return serAttr
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

func GetCertPath() string {
	path, _ := utils.GetMyPath()
	return filepath.Join(path, "certs", "utm.crt")
}

func GetKeyPath() string {
	path, _ := utils.GetMyPath()
	return filepath.Join(path, "certs", "utm.key")
}

func GetCaPath() string {
	path, _ := utils.GetMyPath()
	return filepath.Join(path, "certs", "ca.crt")
}
