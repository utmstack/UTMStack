package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/utmstack/UTMStack/agent/collector"
	"github.com/utmstack/UTMStack/agent/config"
)

var enableTLS bool

var enableIntegrationCmd = &cobra.Command{
	Use:   "enable-integration <integration> [protocol]",
	Short: "Enable integration for a specific integration and protocol",
	Long: `Enable integration for a specific integration and protocol.

For syslog integrations, protocol (tcp/udp) is required.
For file integrations (nginx, postgresql), protocol is not needed.

Available flag: --tls (enable TLS for TCP only)

Examples:
  utmstack_agent enable-integration syslog tcp --tls       # Enable syslog with TLS
  utmstack_agent enable-integration syslog tcp             # Enable syslog without TLS
  utmstack_agent enable-integration nginx                  # Enable file-based nginx collector`,
	Args:    cobra.RangeArgs(1, 2),
	PreRunE: requireInstalled,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Changing integration status ...")
		integration := args[0]

		// Check integration type
		intType := config.ValidateModuleType(integration)

		switch intType {
		case "file":
			// File-based integration - no protocol needed
			paths, err := collector.ChangeFileIntegrationStatus(integration, true)
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
			fmt.Printf("Integration %s enabled. Watching paths:\n", integration)
			for _, p := range paths {
				fmt.Printf("  - %s\n", p)
			}
			time.Sleep(5 * time.Second)
			return nil

		case "syslog":
			// Syslog integration - protocol required
			if len(args) < 2 {
				fmt.Println("Error: protocol (tcp/udp) is required for syslog integrations")
				os.Exit(1)
			}
			proto := args[1]

			var port string
			var err error

			if enableTLS {
				port, err = collector.ChangeIntegrationStatus(integration, proto, true, true)
			} else {
				port, err = collector.ChangeIntegrationStatus(integration, proto, true, false)
			}

			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}

			if enableTLS {
				fmt.Printf("Integration %s %s enabled with TLS on port %s\n", integration, proto, port)
			} else {
				fmt.Printf("Integration %s %s enabled on port %s\n", integration, proto, port)
			}
			time.Sleep(5 * time.Second)
			return nil

		case "netflow":
			// Netflow integration
			if len(args) < 2 {
				fmt.Println("Error: protocol (udp) is required for netflow integration")
				os.Exit(1)
			}
			proto := args[1]
			port, err := collector.ChangeIntegrationStatus(integration, proto, true)
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
			fmt.Printf("Integration %s %s enabled on port %s\n", integration, proto, port)
			time.Sleep(5 * time.Second)
			return nil

		default:
			fmt.Printf("Error: invalid integration: %s\n", integration)
			fmt.Println("Valid syslog integrations:", strings.Join(getSyslogIntegrations(), ", "))
			fmt.Println("Valid file integrations:", strings.Join(getFileIntegrations(), ", "))
			os.Exit(1)
			return nil
		}
	},
}

func getSyslogIntegrations() []string {
	return []string{"syslog", "vmware-esxi", "antivirus-esmc-eset", "antivirus-kaspersky",
		"firewall-cisco-asa", "firewall-cisco-firepower", "cisco-switch", "firewall-meraki",
		"firewall-fortigate-traffic", "firewall-paloalto", "firewall-mikrotik", "firewall-sophos-xg",
		"firewall-sonicwall", "deceptive-bytes", "antivirus-sentinel-one", "ibm-aix",
		"firewall-pfsense", "firewall-fortiweb", "suricata"}
}

func getFileIntegrations() []string {
	return []string{"nginx", "postgresql"}
}

func init() {
	enableIntegrationCmd.Flags().BoolVar(&enableTLS, "tls", false, "Enable TLS for TCP")
	rootCmd.AddCommand(enableIntegrationCmd)
}
