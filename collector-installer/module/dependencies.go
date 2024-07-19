package module

import (
	"fmt"
	"os"
	"runtime"

	"github.com/utmstack/UTMStack/collector-installer/config"
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

func HandleDependenciesPostDownload(collectorType config.Collector) error {
	zipName := config.GetDownloadFilePath(config.DependZipLabel, collectorType)
	if zipName != "" {
		if err := utils.Unzip(zipName, utils.GetMyPath()); err != nil {
			return fmt.Errorf("error unzipping dependencies: %v", err)
		}
		if err := os.Remove(zipName); err != nil {
			fmt.Printf("error deleting dependencies file: %v\n", err)
		}
	}

	return nil
}
