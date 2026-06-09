package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/utmstack/UTMStack/collectors/forwarder/config"
)

var listDataTypesCmd = &cobra.Command{
	Use:   "list-datatypes",
	Short: "List all known DataTypes (built-in and user-defined)",
	Long: `List all DataTypes available to the collector.

Built-in types are shipped with the collector binary.
User-defined types are created automatically when you enable an unknown integration name.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		entries := config.ListAllDataTypes()
		if len(entries) == 0 {
			fmt.Println("No DataTypes found.")
			return nil
		}

		fmt.Printf("%-30s  %-10s  %-9s\n", "NAME", "KIND", "ORIGIN")
		fmt.Printf("%-30s  %-10s  %-9s\n",
			"------------------------------",
			"----------",
			"---------",
		)
		for _, e := range entries {
			fmt.Printf("%-30s  %-10s  %-9s\n", e.Name, e.Kind, e.Origin)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listDataTypesCmd)
}
