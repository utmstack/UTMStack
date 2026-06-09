package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/utmstack/UTMStack/agent/serv"
)

var rootCmd = &cobra.Command{
	Use:   "utmstack_agent",
	Short: "UTMStack Agent CLI",
	Long: `UTMStack Agent CLI

Usage Examples:
  utmstack_agent install <server_address> <utm_key> <skip_cert_validation(yes/no)>
  utmstack_agent change-retention <new_retention>
  utmstack_agent clean-logs
  utmstack_agent uninstall

Note:
  - Make sure to run commands with appropriate permissions.
  - All commands require administrative privileges.
  - For detailed logs, check the service log file.`,
	Run: func(cmd *cobra.Command, args []string) {
		serv.RunService()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
