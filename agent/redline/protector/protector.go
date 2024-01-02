package protector

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/quantfall/holmes"
	"github.com/utmstack/UTMStack/agent/redline/constants"
	"github.com/utmstack/UTMStack/agent/redline/utils"
)

func ProtectService(servName, lockName string, h *holmes.Logger) {
	// Get current path
	path, err := utils.GetMyPath()
	if err != nil {
		fmt.Printf("Failed to get current path: %v", err)
		h.FatalError("Failed to get current path: %v", err)
	}

	bin := constants.GetAgentBin()

	attempts := 0
	for {
		if attempts >= 3 {
			h.Info("%s service has been stopped", servName)
			if err := utils.Execute(filepath.Join(path, bin), path, "send-log", fmt.Sprintf("%s service has been stopped", servName)); err != nil {
				h.Error("error checking %s: error sending log : %v", servName, err)
				time.Sleep(time.Second * 5)
				continue
			}
			if err := utils.RestartService(servName); err != nil {
				h.Error("error checking %s: error restarting %s service: %v", servName, servName, err)
				time.Sleep(time.Second * 5)
				continue
			}

			h.Info("%s restarted correctly", servName)
			time.Sleep(time.Second * 5)
			attempts = 0
			continue
		}

		if isRunning, err := utils.CheckIfServiceIsActive(servName); err != nil {
			h.Error("error checking if %s is running: %v", servName, err)
			time.Sleep(time.Second * 5)
		} else if isRunning {
			time.Sleep(time.Second * 5)
			continue
		}
		if !utils.CheckIfPathExist(filepath.Join(path, "locks", lockName)) {
			attempts++
			time.Sleep(time.Second * 30)
			continue
		}
		if lockName == "utmstack_updater.lock" && runtime.GOOS == "linux" {
			err = utils.Execute(filepath.Join(path, "utmstack_updater_self"), path)
			if err != nil {
				h.Error("error executing utmstack_updater_self: %v", err)
			}
			time.Sleep(time.Second * 5)
			continue
		}
	}
}
