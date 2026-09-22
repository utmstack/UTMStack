package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/utmstack/UTMStack/collectors/forwarder/serv"
)

var rootCmd = &cobra.Command{
	Use:   "utmstack_forwarder",
	Short: "UTMStack Forwarder CLI",
	Long: `UTMStack Forwarder CLI

Usage Examples:
  utmstack_forwarder install <server_address> <utm_key> <skip_cert_validation(yes/no)>
  utmstack_forwarder uninstall
  utmstack_forwarder enable-integration <name> <port> <protocol>
  utmstack_forwarder disable-integration <name> <protocol>
  utmstack_forwarder change-port <name> <protocol> <port>
  utmstack_forwarder change-paths <name> <path1> [path2 ...]
  utmstack_forwarder show-paths <name>
  utmstack_forwarder load-tls-certs <cert> <key> [ca]
  utmstack_forwarder list-datatypes
  utmstack_forwarder remove-datatype <name>

Note:
  - Make sure to run commands with appropriate permissions.
  - All commands require administrative privileges.
  - For detailed logs, check the collector log file.`,
	Run: func(cmd *cobra.Command, args []string) {
		serv.RunService()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
