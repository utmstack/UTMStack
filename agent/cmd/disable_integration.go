package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/utmstack/UTMStack/agent/collector"
	"github.com/utmstack/UTMStack/agent/config"
)

var disableIntegrationCmd = &cobra.Command{
	Use:   "disable-integration <integration> [protocol]",
	Short: "Disable integration for a specific integration and protocol",
	Long: `Disable integration for a specific integration and protocol.

For syslog integrations, protocol (tcp/udp) is required.
For file integrations (nginx, postgresql), protocol is not needed.

Examples:
  utmstack_agent disable-integration syslog tcp             # Disable syslog TCP
  utmstack_agent disable-integration nginx                  # Disable file-based nginx collector`,
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
			_, err := collector.ChangeFileIntegrationStatus(integration, false)
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
			fmt.Printf("Integration %s disabled\n", integration)
			time.Sleep(5 * time.Second)
			return nil

		case "syslog", "netflow":
			// Syslog/netflow integration - protocol required
			if len(args) < 2 {
				fmt.Println("Error: protocol (tcp/udp) is required for this integration type")
				os.Exit(1)
			}
			proto := args[1]

			port, err := collector.ChangeIntegrationStatus(integration, proto, false)
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}

			fmt.Printf("Integration %s %s disabled (port %s freed)\n", integration, proto, port)
			time.Sleep(5 * time.Second)
			return nil

		default:
			fmt.Printf("Error: invalid integration: %s\n", integration)
			os.Exit(1)
			return nil
		}
	},
}

func init() {
	rootCmd.AddCommand(disableIntegrationCmd)
}
