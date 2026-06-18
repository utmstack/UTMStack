package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	collectorpkg "github.com/utmstack/UTMStack/collectors/forwarder/collector"
)

var changePathsCmd = &cobra.Command{
	Use:   "change-paths <name> <path1> [path2 ...]",
	Short: "Change the file paths watched by a file-based integration",
	Long: `Change the file paths monitored by a file-based log integration.
At least one path must be provided.

Examples:
  utmstack_forwarder change-paths nginx /var/log/nginx/access.log
  utmstack_forwarder change-paths postgresql /var/log/postgresql/postgresql.log`,
	Args:    cobra.MinimumNArgs(2),
	PreRunE: requireInstalled,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		paths := args[1:]

		fmt.Printf("Changing paths for integration %q to: %s ...\n", name, strings.Join(paths, ", "))

		result, err := collectorpkg.ChangeFilePaths(name, paths)
		if err != nil {
			fmt.Printf("Error changing paths: %v\n", err)
			os.Exit(1)
		}

		if len(result.OldPaths) > 0 {
			fmt.Printf("Old paths: %s\n", strings.Join(result.OldPaths, ", "))
		}
		fmt.Printf("Paths updated for integration %q\n", name)
		time.Sleep(2 * time.Second)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(changePathsCmd)
}
