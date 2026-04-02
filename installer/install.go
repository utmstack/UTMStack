package main

import (
	"fmt"
	"time"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/docker"
	"github.com/utmstack/UTMStack/installer/setup"
	"github.com/utmstack/UTMStack/installer/updater"
	"github.com/utmstack/UTMStack/installer/utils"
)

func Install() error {
	fmt.Println("### Installing UTMStack ###")

	go updater.MonitorConnection(config.GetCMServer(), 30*time.Second, 3, &config.ConnectedToInternet)

	isInstalledAlready, err := utils.CheckIfServiceIsInstalled("UTMStackComponentsUpdater")
	if err != nil {
		return fmt.Errorf("error checking if service is installed: %v", err)
	}

	if isInstalledAlready {
		fmt.Println("UTMStack is already installed. If you want to re-install it, please remove the service UTMStackComponentsUpdater first.")
		if err := utils.RestartService("UTMStackComponentsUpdater"); err != nil {
			return fmt.Errorf("error restarting service: %v", err)
		}
		return nil
	}

	version, err := updater.GetVersion()
	if err != nil {
		return err
	}

	pass, err := setup.Apply(version.Version, false)
	if err != nil {
		return fmt.Errorf("error applying setup: %v", err)
	}

	fmt.Print("Installing Updater Service")
	updater.InstallService()
	fmt.Println(" [OK]")

	fmt.Println("Running post installation cleanup...")
	if err := docker.PostInstallation(); err != nil {
		fmt.Printf("Warning: post-installation cleanup failed: %v\n", err)
	}

	fmt.Println("Installation finished successfully. We have generated a configuration file for you, please do not modify or remove it. You can find it at /root/utmstack.yml.")
	fmt.Println("You can also use it to re-install your stack in case of a disaster or changes in your hardware. Just run the installer again.")
	fmt.Println("You can access to your Web-GUI at https://<your-server-ip> using admin as your username")
	fmt.Printf("Web-GUI default password for admin: %s \n", pass)
	fmt.Println("You can also access to your Web-based Administration Interface at https://<your-server-ip>:9090 using your Linux system credentials.")
	fmt.Println("Detailed installation logs can be found at /var/log/utmstack-installer.log")

	fmt.Println("### Thanks for using UTMStack ###")

	return nil
}
