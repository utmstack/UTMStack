package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/utmstack/UTMStack/as400/updater/config"
	"github.com/utmstack/UTMStack/as400/updater/service"
	"github.com/utmstack/UTMStack/as400/updater/utils"
)

func main() {
	path := utils.GetMyPath()
	utils.InitLogger(filepath.Join(path, "logs", config.SERV_LOG))

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			fmt.Println("Installing UTMStack AS400 Updater service...")

			fmt.Print(("Creating service ... "))
			service.InstallService()
			fmt.Println("[OK]")

			fmt.Println("UTMStack AS400 Updater service installed correctly")
			return
		case "uninstall":
			fmt.Println("Uninstalling UTMStack AS400 Updater service...")
			service.UninstallService()
			fmt.Println("Service uninstalled successfully")
			return
		case "start":
			fmt.Println("Starting UTMStack AS400 Updater service...")
			return
		case "stop":
			fmt.Println("Stopping UTMStack AS400 Updater service...")
			// Will be handled by systemd
			return
		}
	}

	service.RunService()
}
