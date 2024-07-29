package config

import "runtime"

const (
	SERV_LOG  = "utmstack_updater_self.log"
	SERV_NAME = "UTMStackAgent"
)

func GetAgentBin(typ string) string {
	var bin string = "utmstack_agent_service"
	if typ == "new" {
		bin = bin + "_new"
	}
	if runtime.GOOS == "windows" {
		bin = bin + ".exe"
	}
	return bin
}
