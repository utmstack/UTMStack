package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/utmstack/UTMStack/collectors/forwarder/config"
	"github.com/utmstack/UTMStack/collectors/forwarder/serv"
)

var uninstallCmd = &cobra.Command{
	Use:     "uninstall",
	Short:   "Uninstall the UTMStackForwarder service",
	Args:    cobra.NoArgs,
	PreRunE: requireInstalled,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Uninstalling UTMStackForwarder service ...")

		os.Remove(config.ConfigurationFile)

		serv.UninstallService()

		fmt.Println("[OK]")
		fmt.Println("UTMStackForwarder service uninstalled correctly")
		os.Exit(0)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

