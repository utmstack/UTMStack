package utils

func StopService(name string) error {
	path := GetMyPath()
	err := Execute("systemctl", path, "stop", name)
	if err != nil {
		return Logger.ErrorF("error stopping service: %v", err)
	}
	return nil
}

func UninstallService(name string) error {
	path := GetMyPath()
	err := Execute("systemctl", path, "disable", name)
	if err != nil {
		return Logger.ErrorF("error uninstalling service: %v", err)
	}
	err = Execute("rm", "/etc/systemd/system/", "/etc/systemd/system/"+name+".service")
	if err != nil {
		return Logger.ErrorF("error uninstalling service: %v", err)
	}
	return nil
}

func CheckIfServiceIsInstalled(serv string) (bool, error) {
	path := GetMyPath()
	err := Execute("systemctl", path, "status", serv)
	return err == nil, nil
}
