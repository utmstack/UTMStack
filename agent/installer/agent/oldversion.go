package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/agent/runner/utils"
)

func CleanOldAgent(h *logger.Logger) {
	servName := "utmstack"
	if isOldInstalled, err := utils.CheckIfServiceIsInstalled(servName); err != nil {
		fmt.Printf("error checking utmstack service: %v", err)
		h.ErrorF("error checking utmstack service: %v", err)
	} else if isOldInstalled {
		fmt.Println("Uninstalling UTMStack Agent old version...")
		h.Info("Uninstalling UTMStack Agent old version...")

		// Stopping and uninstalling UTMStack Agent old version
		err := utils.StopService(servName)
		if err != nil {
			h.ErrorF("error stopping %s: %v", servName, err)
		}

		err = utils.UninstallService(servName)
		if err != nil {
			h.ErrorF("error uninstalling %s: %v", servName, err)
		}

		err = stopWazuh()
		if err != nil {
			h.ErrorF("error stopping wazuh: %v", err)
		}

		// Stopping and uninstalling beats
		switch runtime.GOOS {
		case "windows":
			err := stopWinlogbeat()
			if err != nil {
				h.ErrorF("%v", err)
			}
			pathOld := "C:\\Program Files\\UTMStack\\UTMStackAgent"
			err = os.RemoveAll(pathOld)
			if err != nil {
				h.ErrorF("error deleting old agent folder: %v", err)
			}

		case "linux":
			err := uninstallFilebeat()
			if err != nil {
				h.ErrorF("%v", err)
			}
			pathOld := "/opt/linux-agent"
			err = os.RemoveAll(pathOld)
			if err != nil {
				h.ErrorF("error deleting old agent folder: %v", err)
			}
		}
	}
}

func stopWazuh() error {
	servName := "WazuhSvc"
	if isInstalled, err := utils.CheckIfServiceIsInstalled(servName); err != nil {
		return fmt.Errorf("error checking UTMStackAgent service: %v", err)
	} else if isInstalled {
		err := utils.StopService(servName)
		if err != nil {
			return fmt.Errorf("error stopping %s: %v", servName, err)
		}
		err = utils.UninstallService(servName)
		if err != nil {
			return fmt.Errorf("error uninstalling %s: %v", servName, err)
		}
	}

	return nil
}

func stopWinlogbeat() error {
	isRunning, err := utils.IsProcessRunning("winlogbeat")
	if err != nil {
		return fmt.Errorf("error checking if winlogbeat is running: %v", err)
	} else if isRunning {
		err = utils.StopProcess("winlogbeat")
		if err != nil {
			return fmt.Errorf("error stopping winlogbeat: %v", err)
		}
	}

	return nil
}

func uninstallFilebeat() error {
	path, err := utils.GetMyPath()
	if err != nil {
		return fmt.Errorf("error getting current path: %v", err)
	}

	err = utils.Execute("systemctl", path, "stop", "filebeat")
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	family, err := utils.DetectLinuxFamily()
	if err != nil {
		return err
	}

	switch family {
	case "debian":
		err = utils.Execute("apt-get", path, "remove", "--purge", "-y", "filebeat")
		if err != nil {
			return fmt.Errorf("%s", err)
		}
	case "rhel":
		err = utils.Execute("systemctl", filepath.Join(path, "beats", "filebeat"), "stop", "filebeat")
		if err != nil {
			return fmt.Errorf("%s", err)
		}
		err = utils.Execute("systemctl", filepath.Join(path, "beats", "filebeat"), "disable", "filebeat")
		if err != nil {
			return fmt.Errorf("%s", err)
		}
		err = utils.Execute("echo", filepath.Join(path, "beats", "filebeat"), "y", "|", "yum", "remove", "filebeat")
		if err != nil {
			return fmt.Errorf("%s", err)
		}

	}

	err = utils.Execute("rm", path, "-rf", "/etc/filebeat")
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	return nil
}
