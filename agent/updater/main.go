package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/utmstack/UTMStack/agent/updater/config"
	"github.com/utmstack/UTMStack/agent/updater/service"
	"github.com/utmstack/UTMStack/agent/updater/utils"
)

func main() {
	path := utils.GetMyPath()
	utils.InitLogger(filepath.Join(path, "logs", config.SERV_LOG))

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			fmt.Println("Installing UTMStack Updater service...")

			fmt.Print(("Creating service ... "))
			service.InstallService()
			fmt.Println("[OK]")

			fmt.Println("UTMStackUpdater service installed correctly")
			return
		case "uninstall":
			fmt.Println("Uninstalling UTMStack Updater service...")
			service.UninstallService()
			fmt.Println("Service uninstalled successfully")
			return
		case "start":
			fmt.Println("Starting UTMStack Updater service...")
			return
		case "stop":
			fmt.Println("Stopping UTMStack Updater service...")
			// Will be handled by systemd
			return
		}
	}

	service.RunService()
}
