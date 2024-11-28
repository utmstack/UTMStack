package utils

import (
	"fmt"
)

func StopService(name string) error {
	return RunCmd("systemctl", "stop", name)
}

func UninstallService(name string) error {
	err := RunCmd("systemctl", "disable", name)
	if err != nil {
		return fmt.Errorf("error uninstalling service: %v", err)
	}
	err = RunCmd("rm", "/etc/systemd/system/"+name+".service")
	if err != nil {
		return fmt.Errorf("error uninstalling service: %v", err)
	}

	return nil
}

func CheckIfServiceIsInstalled(serv string) (bool, error) {
	err := RunCmd("systemctl", "status", serv)
	if err != nil {
		return false, nil
	}
	return true, nil
}
