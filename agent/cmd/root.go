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
  utmstack_agent enable-integration <integration> <protocol> [--tls]
  utmstack_agent disable-integration <integration> <protocol>
  utmstack_agent change-port <integration> <protocol> <new_port>
  utmstack_agent change-retention <new_retention>
  utmstack_agent load-tls-certs <cert> <key> [ca]
  utmstack_agent clean-logs
  utmstack_agent uninstall

TLS Certificate Management:
  # Load your own certificates (RECOMMENDED)
  utmstack_agent load-tls-certs /path/to/server.crt /path/to/server.key /path/to/ca.crt
  utmstack_agent load-tls-certs /path/to/server.crt /path/to/server.key  # Without CA

TLS Integration Examples:
  utmstack_agent enable-integration syslog tcp --tls       # Enable with TLS
  utmstack_agent enable-integration syslog tcp             # Enable without TLS (default)
  utmstack_agent disable-integration syslog tcp            # Disable (auto-disables TLS)

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
