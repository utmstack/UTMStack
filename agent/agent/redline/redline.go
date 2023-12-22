package redline

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/quantfall/holmes"
	"github.com/utmstack/UTMStack/agent/agent/configuration"
	"github.com/utmstack/UTMStack/agent/agent/utils"
)

func CheckRedlineService(h *holmes.Logger) {
	// Get current path
	path, err := utils.GetMyPath()
	if err != nil {
		fmt.Printf("Failed to get current path: %v", err)
		h.FatalError("Failed to get current path: %v", err)
	}

	bin := configuration.GetAgentBin()

	time.Sleep(5 * time.Minute)
	attempts := 0
	for {
		if attempts >= 3 {
			h.Info("Redline service has been stopped")
			if err := utils.Execute(filepath.Join(path, bin), path, "send-log", fmt.Sprintf("%s service has been stopped", configuration.RedlineServName)); err != nil {
				h.Error("error checking %s: error sending log : %v", configuration.RedlineServName, err)
				time.Sleep(time.Second * 5)
				continue
			}
			if err := utils.RestartService(configuration.RedlineServName); err != nil {
				h.Error("error restarting %s service: %v", configuration.RedlineServName, err)
				time.Sleep(time.Second * 5)
				continue
			}

			h.Info("%s restarted correctly", configuration.RedlineServName)
			time.Sleep(time.Second * 5)
			attempts = 0
			continue
		}

		if isRunning, err := utils.CheckIfServiceIsActive(configuration.RedlineServName); err != nil {
			h.Error("error checking if %s is running: %v", configuration.RedlineServName, err)
			time.Sleep(time.Second * 5)
		} else if isRunning {
			time.Sleep(time.Second * 5)
			continue
		}
		if !utils.CheckIfPathExist(filepath.Join(path, "locks", configuration.RedlineLockName)) {
			attempts++
			time.Sleep(time.Second * 30)
		}
	}
}
