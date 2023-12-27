package constants

import "runtime"

const (
	SERV_NAME         = "UTMStackUpdaterSelf"
	SERV_LOG          = "utmstack_updater_self.log"
	UPDATER_SERV_NAME = "UTMStackUpdater"
	UPDATER_SERV_LOCK = "utmstack_updater.lock"
	Bucket            = "https://cdn.utmstack.com/agent_updates/"
	SERV_LOCK         = "utmstack_updater.lock"
)

func GetUpdaterBin() string {
	var bin string
	switch runtime.GOOS {
	case "windows":
		bin = "utmstack_updater_service.exe"
	case "linux":
		bin = "utmstack_updater_service"
	}
	return bin
}
