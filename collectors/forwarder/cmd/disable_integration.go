package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	collectorpkg "github.com/utmstack/UTMStack/collectors/forwarder/collector"
)

var disableIntegrationCmd = &cobra.Command{
	Use:   "disable-integration <name> <protocol>",
	Short: "Disable a log integration",
	Long: `Disable a log integration by name and protocol.

Protocol: tcp | udp | tls | http | https

Examples:
  utmstack_forwarder disable-integration firewall-cisco-asa udp
  utmstack_forwarder disable-integration firewall-cisco-asa tcp
  utmstack_forwarder disable-integration my-app http`,
	Args:    cobra.ExactArgs(2),
	PreRunE: requireInstalled,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		proto := strings.ToLower(args[1])

		switch proto {
		case "tcp", "udp", "tls", "http", "https":
		default:
			return fmt.Errorf("invalid protocol %q: must be tcp, udp, tls, http or https", proto)
		}

		fmt.Printf("Disabling %s integration %q...\n", proto, name)

		switch proto {
		case "http", "https":
			if err := collectorpkg.DisableHTTPIntegration(name, proto); err != nil {
				return fmt.Errorf("error disabling integration: %w", err)
			}
		default:
			useTLS := proto == "tls"
			networkProto := proto
			if useTLS {
				networkProto = "tcp"
			}
			port, err := collectorpkg.ChangeIntegrationStatus(name, networkProto, false, useTLS)
			if err != nil {
				return fmt.Errorf("error disabling integration: %w", err)
			}
			fmt.Printf("Integration %q disabled on %s port %s\n", name, proto, port)
		}

		time.Sleep(2 * time.Second)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(disableIntegrationCmd)
}
