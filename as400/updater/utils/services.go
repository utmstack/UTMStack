package utils

import (
	"fmt"
	"strings"
)

func CheckIfServiceIsActive(serv string) (bool, error) {
	path := GetMyPath()
	output, errB := ExecuteWithResult("systemctl", path, "is-active", serv)
	if errB {
		return false, nil
	}
	serviceStatus := strings.ToLower(strings.TrimSpace(output))
	return serviceStatus == "active", nil
}

func RestartService(serv string) error {
	path := GetMyPath()
	isRunning, err := CheckIfServiceIsActive(serv)
	if err != nil {
		return fmt.Errorf("error checking if %s service is active: %v", serv, err)
	}

	if isRunning {
		err := Execute("systemctl", path, "restart", serv)
		if err != nil {
			return fmt.Errorf("error restarting service: %v", err)
		}
	} else {
		err := Execute("systemctl", path, "start", serv)
		if err != nil {
			return fmt.Errorf("error starting service: %v", err)
		}
	}
	return nil
}

func StopService(name string) error {
	path := GetMyPath()
	err := Execute("systemctl", path, "stop", name)
	if err != nil {
		return fmt.Errorf("error stopping service: %v", err)
	}
	return nil
}

func StartService(name string) error {
	path := GetMyPath()
	err := Execute("systemctl", path, "start", name)
	if err != nil {
		return fmt.Errorf("error starting service: %v", err)
	}
	return nil
}
