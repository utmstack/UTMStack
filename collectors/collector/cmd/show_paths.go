package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	collectorpkg "github.com/utmstack/UTMStack/collectors/forwarder/collector"
)

var showPathsCmd = &cobra.Command{
	Use:   "show-paths <name>",
	Short: "Show the file paths watched by a file-based integration",
	Long: `Display the file paths currently monitored by a file-based log integration.

Examples:
  utmstack_forwarder show-paths nginx
  utmstack_forwarder show-paths postgresql`,
	Args:    cobra.ExactArgs(1),
	PreRunE: requireInstalled,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		paths, err := collectorpkg.GetFileIntegrationPaths(name)
		if err != nil {
			fmt.Printf("Error retrieving paths: %v\n", err)
			os.Exit(1)
		}

		if len(paths) == 0 {
			fmt.Printf("No paths configured for integration %q\n", name)
			return nil
		}

		fmt.Printf("Paths for integration %q:\n", name)
		for _, p := range paths {
			fmt.Printf("  - %s\n", p)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(showPathsCmd)
}
