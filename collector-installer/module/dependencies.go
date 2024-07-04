package module

import (
	"runtime"

	"github.com/utmstack/UTMStack/collector-installer/utils"
)

func InstallJava() error {
	if runtime.GOOS == "linux" {
		err := utils.Execute("apt-get", "", "update")
		if err != nil {
			return err
		}

		err = utils.Execute("apt", "", "install", "-y", "openjdk-11-jdk")
		if err != nil {
			return err
		}
	}
	return nil
}

func UninstallJava() error {
	if runtime.GOOS == "linux" {
		err := utils.Execute("apt", "", "remove", "-y", "openjdk-11-jdk")
		if err != nil {
			return err
		}
		err = utils.Execute("apt", "", "autoremove", "-y")
		if err != nil {
			return err
		}
	}
	return nil
}
