package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent/installer/config"
	"github.com/utmstack/UTMStack/agent/installer/updates"
	"github.com/utmstack/UTMStack/agent/installer/utils"
)

func main() {
	path := utils.GetMyPath()
	utils.InitBeautyLogger("UTMStack Agent Installer", filepath.Join(utils.GetMyPath(), "logs", config.INSTALLER_LOG_FILE))
	utils.Logger.PrintBanner()

	if len(os.Args) > 1 {
		mode := os.Args[1]
		switch mode {
		case "install":
			ip, utmKey, skip := os.Args[2], os.Args[3], os.Args[4]

			if strings.Count(utmKey, "*") == len(utmKey) {
				utils.Logger.WriteFatal("the connection key provided is incorrect. Please make sure you use the 'copy' icon from the integrations section to get the value of the masked key value.")
			}

			utils.Logger.WriteSimpleMessage("Installing UTMStack Agent...")
			if !utils.IsPortOpen(ip, config.AgentManagerPort) || !utils.IsPortOpen(ip, config.LogAuthProxyPort) {
				utils.Logger.WriteFatal("one or more of the requiered ports are closed. Please open ports 9000 and 50051.")
			}

			certsPath := filepath.Join(path, "certs")
			err := utils.CreatePathIfNotExist(certsPath)
			if err != nil {
				utils.Logger.WriteFatal("error creating path: %s", err)
			}

			err = utils.GenerateCerts(certsPath)
			if err != nil {
				utils.Logger.WriteFatal("error generating certificates: %s", err)
			}

			utils.Logger.WriteSimpleMessage("Downloading UTMStack dependencies...")
			err = updates.DownloadDependencies(ip, utmKey, skip)
			if err != nil {
				utils.Logger.WriteFatal("error downloading dependencies", err)
			}
			utils.Logger.WriteSuccessfull("UTMStack dependencies downloaded correctly.")

			utils.Logger.WriteSimpleMessage("Installing service...")
			err = config.ConfigureService(ip, utmKey, skip, "install")
			if err != nil {
				utils.Logger.WriteFatal("error installing UTMStack service", err)
			}

			utils.Logger.WriteSuccessfull("Service installed correctly")
			utils.Logger.WriteSuccessfull("UTMStack Agent installed correctly.")

			time.Sleep(5 * time.Second)
			os.Exit(0)

		case "uninstall":
			utils.Logger.WriteSimpleMessage("Uninstalling UTMStack Agent...")

			if isInstalled, err := utils.CheckIfServiceIsInstalled("UTMStackAgent"); err != nil {
				utils.Logger.WriteFatal("error checking UTMStackAgent service", err)
			} else if isInstalled {
				utils.Logger.WriteSimpleMessage("Uninstalling UTMStack service...")
				err = config.ConfigureService("", "", "", "uninstall")
				if err != nil {
					utils.Logger.WriteFatal("error uninstalling UTMStack service", err)
				}

				utils.Logger.WriteSuccessfull("UTMStack service uninstalled correctly.")
				time.Sleep(5 * time.Second)
				os.Exit(0)

			} else {
				utils.Logger.WriteFatal("UTMStackAgent not installed", nil)
			}

		default:
			utils.Logger.WriteError("unknown option", nil)
		}
	}
}
