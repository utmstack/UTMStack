package configuration

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
	CHECK_EVERY           = 5 * time.Minute
)

type ServiceAttribt struct {
	ServName string
	ServBin  string
	ServLock string
}

// GetServAttr returns a map of ServiceAttribt
func GetServAttr() map[string]ServiceAttribt {
	serAttr := map[string]ServiceAttribt{
		"agent":   {ServName: "UTMStackAgent", ServBin: "utmstack_agent_service", ServLock: "utmstack_agent.lock"},
		"updater": {ServName: "UTMStackUpdater", ServBin: "utmstack_updater_self", ServLock: "utmstack_updater.lock"},
		"redline": {ServName: "UTMStackRedline", ServBin: "utmstack_redline_service", ServLock: "utmstack_redline.lock"},
	}

	switch runtime.GOOS {
	case "windows":
		for code, att := range serAttr {
			att.ServBin += ".exe"
			serAttr[code] = att
		}
	}

	return serAttr
}

// GetAgentBin returns the agent binary name
func GetAgentBin() string {
	bin := "utmstack_agent_service"
	if runtime.GOOS == "windows" {
		bin = bin + ".exe"
	}
	return bin
}

// GetCertPath returns the path to the certificate
func GetCertPath() string {
	path, _ := utils.GetMyPath()
	return filepath.Join(path, "certs", "utm.crt")
}

// GetKeyPath returns the path to the key
func GetKeyPath() string {
	path, _ := utils.GetMyPath()
	return filepath.Join(path, "certs", "utm.key")
}

// GetCaPath returns the path to the CA
func GetCaPath() string {
	path, _ := utils.GetMyPath()
	return filepath.Join(path, "certs", "ca.crt")
}
