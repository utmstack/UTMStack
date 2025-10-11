package main

import (
	"fmt"
	"os"
	"time"

	pb "github.com/utmstack/UTMStack/utmstack-collector/agent"
	"github.com/utmstack/UTMStack/utmstack-collector/config"
	"github.com/utmstack/UTMStack/utmstack-collector/database"
	"github.com/utmstack/UTMStack/utmstack-collector/logservice"
	"github.com/utmstack/UTMStack/utmstack-collector/models"
	"github.com/utmstack/UTMStack/utmstack-collector/serv"
	"github.com/utmstack/UTMStack/utmstack-collector/updates"
	"github.com/utmstack/UTMStack/utmstack-collector/utils"
)

func main() {
	utils.InitLogger(config.ServiceLogFile)

	if len(os.Args) > 1 {
		arg := os.Args[1]

		isInstalled, err := utils.CheckIfServiceIsInstalled("UTMStackCollector")
		if err != nil {
			fmt.Println("Error checking if service is installed: ", err)
			os.Exit(1)
		}
		if arg != "install" && !isInstalled {
			fmt.Println("UTMStackCollector service is not installed")
			os.Exit(1)
		} else if arg == "install" && isInstalled {
			fmt.Println("UTMStackCollector service is already installed")
			os.Exit(1)
		}

		switch arg {
		case "run":
			serv.RunService()
		case "install":
			utils.PrintBanner()
			fmt.Println("Installing UTMStackCollector service ...")

			fmt.Println("[OK]")

			cnf, utmKey := config.GetInitialConfig()

			fmt.Print("Checking server connection ... ")
			if err := utils.ArePortsReachable(cnf.Server, config.AgentManagerPort, config.LogAuthProxyPort, config.DependenciesPort); err != nil {
				fmt.Println("\nError trying to connect to server: ", err)
				os.Exit(1)
			}
			fmt.Println("[OK]")

			fmt.Print("Downloading Version ... ")
			if err := updates.DownloadVersion(cnf.Server, cnf.SkipCertValidation); err != nil {
				fmt.Println("\nError downloading version: ", err)
				os.Exit(1)
			}
			fmt.Println("[OK]")

			fmt.Print("Configuring collector ... ")
			err = pb.RegisterCollector(cnf, utmKey)
			if err != nil {
				fmt.Println("\nError registering collector: ", err)
				os.Exit(1)
			}
			if err = config.SaveConfig(cnf); err != nil {
				fmt.Println("\nError saving config: ", err)
				os.Exit(1)
			}

			if err := logservice.SetDataRetention(""); err != nil {
				fmt.Println("\nError setting retention: ", err)
				os.Exit(1)
			}
			fmt.Println("[OK]")

			fmt.Print(("Creating service ... "))
			serv.InstallService()
			fmt.Println("[OK]")
			fmt.Println("UTMStackCollector service installed correctly")

		case "change-retention":
			fmt.Println("Changing log retention ...")
			retention := os.Args[2]

			if err := logservice.SetDataRetention(retention); err != nil {
				fmt.Println("Error trying to change retention: ", err)
				os.Exit(1)
			}

			fmt.Printf("Retention changed correctly to %s\n", retention)
			time.Sleep(5 * time.Second)

		case "clean-logs":
			fmt.Println("Cleaning old logs ...")
			db := database.GetDB()
			datR, err := logservice.GetDataRetention()
			if err != nil {
				fmt.Println("Error getting retention: ", err)
				os.Exit(1)
			}
			_, err = db.DeleteOld(models.Log{}, datR)
			if err != nil {
				fmt.Println("Error cleaning logs: ", err)
				os.Exit(1)
			}
			fmt.Println("Logs cleaned correctly")
			time.Sleep(5 * time.Second)

		case "uninstall":
			fmt.Println("Uninstalling UTMStackCollector service ...")

			cnf, err := config.GetCurrentConfig()
			if err != nil {
				fmt.Println("Error getting config: ", err)
				os.Exit(1)
			}
			if err = pb.DeleteAgent(cnf); err != nil {
				utils.Logger.ErrorF("error deleting collector: %v", err)
			}

			os.Remove(config.ConfigurationFile)

			serv.UninstallService()

			fmt.Println("[OK]")
			fmt.Println("UTMStackCollector service uninstalled correctly")
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
	fmt.Println("### UTMStack Collector ###")
	fmt.Println("Usage:")
	fmt.Println("  To run the service:                     ./utmstack_collector run")
	fmt.Println("  To install the service:                 ./utmstack_collector install")
	fmt.Println("  To change log retention:                ./utmstack_collector change-retention <new_retention>")
	fmt.Println("  To clean old logs:                      ./utmstack_collector clean-logs")
	fmt.Println("  To uninstall the service:               ./utmstack_collector uninstall")
	fmt.Println("  To debug UTMStack installation:         ./utmstack_collector debug-utmstack")
	fmt.Println("  For help (this message):                ./utmstack_collector help")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  run                      Run the UTMStackCollector service")
	fmt.Println("  install                  Install the UTMStackCollector service")
	fmt.Println("  change-retention         Change the log retention to <new_retention>. Retention must be a number of megabytes. Example: 20")
	fmt.Println("  clean-logs               Clean old logs from the database")
	fmt.Println("  uninstall                Uninstall the UTMStackCollector service")
	fmt.Println("  debug-utmstack           Debug UTMStack installation validation")
	fmt.Println("  help                     Display this help message")
	fmt.Println()
	fmt.Println("Requirements:")
	fmt.Println("  - UTMStack must be installed on this system")
	fmt.Println("  - File /utmstack.yaml must exist in root directory")
	fmt.Println("  - Directory /utmstack/ must exist")
	fmt.Println()
	fmt.Println("Note:")
	fmt.Println("  - Make sure to run commands with appropriate permissions.")
	fmt.Println("  - All commands require administrative privileges.")
	fmt.Println("  - For detailed logs, check the service log file.")
	fmt.Println()
	os.Exit(0)
}
