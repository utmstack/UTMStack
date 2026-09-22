package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	collectorpkg "github.com/utmstack/UTMStack/collectors/forwarder/collector"
)

var changePortCmd = &cobra.Command{
	Use:   "change-port <name> <protocol> <port>",
	Short: "Change the listening port for an integration",
	Long: `Change the listening port for a log integration.

Examples:
  utmstack_forwarder change-port cisco-asa tcp 5140
  utmstack_forwarder change-port netflow udp 2055`,
	Args:    cobra.ExactArgs(3),
	PreRunE: requireInstalled,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		proto := args[1]
		port := args[2]

		fmt.Printf("Changing port for integration %q (%s) to %s ...\n", name, proto, port)

		result, err := collectorpkg.ChangePort(name, proto, port)
		if err != nil {
			fmt.Printf("Error changing port: %v\n", err)
			os.Exit(1)
		}

		if result.Warning != "" {
			fmt.Printf("WARNING: %s\n", result.Warning)
		}
		fmt.Printf("Port changed from %s to %s for integration %q (%s)\n", result.OldPort, port, name, proto)
		time.Sleep(2 * time.Second)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(changePortCmd)
}
