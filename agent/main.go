package main

import (
	"fmt"
	"os"
	"time"

	pb "github.com/utmstack/UTMStack/agent/agent"
	"github.com/utmstack/UTMStack/agent/collectors"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/modules"
	"github.com/utmstack/UTMStack/agent/serv"
	"github.com/utmstack/UTMStack/agent/updates"
	"github.com/utmstack/UTMStack/agent/utils"
)

func main() {
	utils.InitLogger(config.ServiceLogFile)

	if len(os.Args) > 1 {
		arg := os.Args[1]

		isInstalled, err := utils.CheckIfServiceIsInstalled("UTMStackAgent")
		if err != nil {
			fmt.Println("Error checking if service is installed: ", err)
			os.Exit(1)
		}
		if arg != "install" && !isInstalled {
			fmt.Println("UTMStackAgent service is not installed")
			os.Exit(1)
		} else if arg == "install" && isInstalled {
			fmt.Println("UTMStackAgent service is already installed")
			os.Exit(1)
		}

		switch arg {
		case "run":
			serv.RunService()
		case "install":
			utils.PrintBanner()
			fmt.Println("Installing UTMStackAgent service ...")

			cnf, utmKey := config.GetInitialConfig()

			fmt.Print("Checking server connection ... ")
			if err := utils.ArePortsReachable(cnf.Server, config.AgentManagerPort, config.LogAuthProxyPort, config.DependenciesPort); err != nil {
				fmt.Println("\nError trying to connect to server: ", err)
				os.Exit(1)
			}
			fmt.Println("[OK]")

			fmt.Print("Downloading dependencies ... ")
			if err := updates.DownloadFirstDependencies(cnf.Server, utmKey, cnf.SkipCertValidation); err != nil {
				fmt.Println("\nError downloading dependencies: ", err)
				os.Exit(1)
			}
			fmt.Println("[OK]")

			fmt.Print("Configuring agent ... ")
			err = pb.RegisterAgent(cnf, utmKey)
			if err != nil {
				fmt.Println("\nError registering agent: ", err)
				os.Exit(1)
			}
			if err = config.SaveConfig(cnf); err != nil {
				fmt.Println("\nError saving config: ", err)
				os.Exit(1)
			}
			if err = modules.ConfigureCollectorFirstTime(); err != nil {
				fmt.Println("\nError configuring collector: ", err)
				os.Exit(1)
			}
			if err = collectors.InstallCollectors(); err != nil {
				fmt.Println("\nError installing collectors: ", err)
				os.Exit(1)
			}
			fmt.Println("[OK]")

			fmt.Print(("Creating service ... "))
			serv.InstallService()
			fmt.Println("[OK]")
			fmt.Println("UTMStackAgent service installed correctly")

		case "enable-integration", "disable-integration":
			fmt.Println("Changing integration status ...")
			integration := os.Args[2]
			proto := os.Args[3]

			port, err := modules.ChangeIntegrationStatus(integration, proto, (arg == "enable-integration"))
			if err != nil {
				fmt.Println("Error trying to change integration status: ", err)
				os.Exit(1)
			}
			fmt.Printf("Action %s %s %s correctly in port %s\n", arg, integration, proto, port)
			time.Sleep(5 * time.Second)

		case "change-port":
			fmt.Println("Changing integration port ...")
			integration := os.Args[2]
			proto := os.Args[3]
			port := os.Args[4]

			old, err := modules.ChangePort(integration, proto, port)
			if err != nil {
				fmt.Println("Error trying to change integration port: ", err)
				os.Exit(1)
			}
			fmt.Printf("Port changed correctly from %s to %s\n", old, port)
			time.Sleep(5 * time.Second)

		case "uninstall":
			fmt.Print("Uninstalling UTMStackAgent service ...")

			cnf, err := config.GetCurrentConfig()
			if err != nil {
				fmt.Println("Error getting config: ", err)
				os.Exit(1)
			}
			if err = pb.DeleteAgent(cnf); err != nil {
				utils.Logger.ErrorF("error deleting agent: %v", err)
			}
			if err = collectors.UninstallCollectors(); err != nil {
				utils.Logger.Fatal("error uninstalling collectors: %v", err)
			}
			os.Remove(config.ConfigurationFile)

			serv.UninstallService()

			fmt.Println("[OK]")
			fmt.Println("UTMStackAgent service uninstalled correctly")
			os.Exit(1)
		case "help":
			Help()
		default:
			fmt.Println("unknown option")
		}
	} else {
		serv.RunService()
	}
}

func Help() {
	fmt.Println("### UTMStackAgent ###")
	fmt.Println("Usage:")
	fmt.Println("  To run the service:                     ./utmstack_agent run")
	fmt.Println("  To install the service:                 ./utmstack_agent install")
	fmt.Println("  To enable integration:                  ./utmstack_agent enable-integration <integration> <protocol>")
	fmt.Println("  To disable integration:                 ./utmstack_agent disable-integration <integration> <protocol>")
	fmt.Println("  To change integration port:             ./utmstack_agent change-port <integration> <protocol> <new_port>")
	fmt.Println("  To uninstall the service:               ./utmstack_agent uninstall")
	fmt.Println("  For help (this message):                ./utmstack_agent help")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  run                      Run the UTMStackAgent service")
	fmt.Println("  install                  Install the UTMStackAgent service")
	fmt.Println("  enable-integration       Enable integration for a specific <integration> and <protocol>")
	fmt.Println("  disable-integration      Disable integration for a specific <integration> and <protocol>")
	fmt.Println("  change-port              Change the port for a specific <integration> and <protocol> to <new_port>")
	fmt.Println("  uninstall                Uninstall the UTMStackAgent service")
	fmt.Println("  help                     Display this help message")
	fmt.Println()
	fmt.Println("Note:")
	fmt.Println("  - Make sure to run commands with appropriate permissions.")
	fmt.Println("  - All commands require administrative privileges.")
	fmt.Println("  - For detailed logs, check the service log file.")
	fmt.Println()
	os.Exit(0)
}
