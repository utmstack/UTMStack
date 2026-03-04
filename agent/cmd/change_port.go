package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/utmstack/UTMStack/agent/collector"
)

var changePortCmd = &cobra.Command{
	Use:     "change-port <integration> <protocol> <new_port>",
	Short:   "Change the port for a specific integration and protocol",
	Args:    cobra.ExactArgs(3),
	PreRunE: requireInstalled,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Changing integration port ...")
		integration := args[0]
		proto := args[1]
		port := args[2]

		result, err := collector.ChangePort(integration, proto, port)
		if err != nil {
			fmt.Println("Error trying to change integration port: ", err)
			os.Exit(1)
		}
		fmt.Printf("Port changed correctly from %s to %s\n", result.OldPort, port)
		if result.Warning != "" {
			fmt.Printf("Warning: %s\n", result.Warning)
		}
		time.Sleep(5 * time.Second)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(changePortCmd)
}
