package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/utmstack/UTMStack/collectors/forwarder/config"
	"github.com/utmstack/UTMStack/collectors/forwarder/serv"
	"github.com/utmstack/UTMStack/collectors/forwarder/upstream"
	"github.com/utmstack/UTMStack/collectors/forwarder/utils"
)

var installCmd = &cobra.Command{
	Use:     "install <server_address> <utm_key> <skip_cert_validation(yes/no)>",
	Short:   "Install the UTMStackForwarder service",
	Args:    cobra.ExactArgs(3),
	PreRunE: requireNotInstalled,
	RunE: func(cmd *cobra.Command, args []string) error {
		cnf := &config.Config{
			Server:             args[0],
			SkipCertValidation: args[2] == "yes",
		}
		utmKey := args[1]

		utils.PrintBanner()
		fmt.Println("Installing UTMStackForwarder service ...")

		fmt.Print("Checking server connection ... ")
		if err := utils.ArePortsReachable(cnf.Server, config.AgentManagerPort, config.LogAuthProxyPort); err != nil {
			fmt.Println("\nError trying to connect to server: ", err)
			os.Exit(1)
		}
		fmt.Println("[OK]")

		fmt.Print("Configuring collector ... ")
		if err := upstream.RegisterCollector(cnf, utmKey); err != nil {
			fmt.Println("\nError registering collector: ", err)
			os.Exit(1)
		}
		if err := config.SaveConfig(cnf); err != nil {
			fmt.Println("\nError saving config: ", err)
			os.Exit(1)
		}
		fmt.Println("[OK]")

		fmt.Print("Creating service ... ")
		serv.InstallService()
		fmt.Println("[OK]")
		fmt.Println("UTMStackForwarder service installed correctly")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
