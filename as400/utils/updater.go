package utils

import "fmt"

func InstallUpdater() error {
	updaterPath := GetMyPath() + "/utmstack_updater_service"

	if err := Execute("chmod", GetMyPath(), "+x", updaterPath); err != nil {
		return fmt.Errorf("error setting execute permissions: %v", err)
	}

	if err := Execute(updaterPath, GetMyPath(), "install"); err != nil {
		return fmt.Errorf("error installing updater service: %v", err)
	}

	return nil
}

func UninstallUpdater() error {
	updaterPath := GetMyPath() + "/utmstack_updater_service"

	if !CheckIfPathExist(updaterPath) {
		return nil
	}

	if err := Execute(updaterPath, GetMyPath(), "uninstall"); err != nil {
		return fmt.Errorf("error uninstalling updater service: %v", err)
	}

	return nil
}
