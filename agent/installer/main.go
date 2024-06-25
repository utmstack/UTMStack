package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent/installer/agent"
	"github.com/utmstack/UTMStack/agent/installer/configuration"
	"github.com/utmstack/UTMStack/agent/installer/utils"
)

func main() {
	beautyLogger := utils.GetBeautyLogger()
	beautyLogger.PrintBanner()

	path := utils.GetMyPath()
	h := utils.CreateLogger(filepath.Join(path, "logs", configuration.INSTALLER_LOG_FILE))

	if len(os.Args) > 1 {
		mode := os.Args[1]
		switch mode {
		case "install":
			ip, utmKey, skip := os.Args[2], os.Args[3], os.Args[4]

			if strings.Count(utmKey, "*") == len(utmKey) {
				beautyLogger.WriteError("The connection key provided is incorrect. Please make sure you use the 'copy' icon from the integrations section to get the value of the masked key value.", nil)
				h.Fatal("The connection key provided is incorrect. Please make sure you use the 'copy' icon from the integrations section to get the value of the masked key value.")
			}

			beautyLogger.WriteSimpleMessage("Installing UTMStack Agent...")
			if !utils.IsPortOpen(ip, configuration.AgentManagerPort) || !utils.IsPortOpen(ip, configuration.LogAuthProxyPort) {
				beautyLogger.WriteError("one or more of the requiered ports are closed. Please open ports 9000 and 50051.", nil)
				h.Fatal("Error installing the UTMStack Agent: one or more of the requiered ports are closed. Please open ports 9000 and 50051.")
			}

			certsPath := filepath.Join(path, "certs")
			err := utils.CreatePathIfNotExist(certsPath)
			if err != nil {
				h.Fatal("error creating path: %s", err)
			}

			err = utils.GenerateCerts(certsPath)
			if err != nil {
				h.Fatal("error generating certificates: %v", err)
			}

			beautyLogger.WriteSimpleMessage("Downloading UTMStack dependencies...")
			err = agent.DownloadDependencies(ip, utmKey, skip, h)
			if err != nil {
				beautyLogger.WriteError("error downloading dependencies", err)
				h.Fatal("error downloading dependencies: %v", err)
			}
			beautyLogger.WriteSuccessfull("UTMStack dependencies downloaded correctly.")

			beautyLogger.WriteSimpleMessage("Installing service...")
			err = agent.ConfigureService(ip, utmKey, skip, "install")
			if err != nil {
				beautyLogger.WriteError("error installing UTMStack services", err)
				h.Fatal("error installing UTMStack services: %v", err)
			}

			beautyLogger.WriteSuccessfull("Services installed correctly")
			beautyLogger.WriteSuccessfull("UTMStack Agent installed correctly.")

			time.Sleep(5 * time.Second)
			os.Exit(0)

		case "uninstall":
			beautyLogger.WriteSimpleMessage("Uninstalling UTMStack Agent...")

			if isInstalled, err := utils.CheckIfServiceIsInstalled("UTMStackAgent"); err != nil {
				beautyLogger.WriteError("error checking UTMStackAgent service", err)
				h.Fatal("error checking UTMStackAgent service: %v", err)
			} else if isInstalled {
				beautyLogger.WriteSimpleMessage("Uninstalling UTMStack services...")
				err = agent.ConfigureService("", "", "", "uninstall")
				if err != nil {
					beautyLogger.WriteError("error uninstalling UTMStack services", err)
					h.Fatal("error uninstalling UTMStack services: %v", err)
				}

				beautyLogger.WriteSuccessfull("UTMStack services uninstalled correctly.")
				time.Sleep(5 * time.Second)
				os.Exit(0)

			} else {
				beautyLogger.WriteError("UTMStackAgent not installed", nil)
				h.Fatal("UTMStackAgent not installed")
			}

		default:
			beautyLogger.WriteError("unknown option", nil)
		}
	}
}
