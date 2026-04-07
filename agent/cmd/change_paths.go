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

var changePathsCmd = &cobra.Command{
	Use:   "change-paths <integration> <path1> [path2] [path3] ...",
	Short: "Change the file paths for a file-based integration",
	Long: `Change the file paths for a file-based integration.

Glob patterns are supported (e.g., /var/log/nginx/*.log).

Examples:
  utmstack_agent change-paths nginx /var/log/nginx/access.log /var/log/nginx/error.log
  utmstack_agent change-paths postgresql "/var/log/postgresql/*.log"`,
	Args:    cobra.MinimumNArgs(2),
	PreRunE: requireInstalled,
	RunE: func(cmd *cobra.Command, args []string) error {
		integration := args[0]
		paths := args[1:]

		// Validate this is a file-type integration
		if config.ValidateModuleType(integration) != "file" {
			fmt.Printf("Error: %s is not a file-based integration\n", integration)
			fmt.Println("Valid file integrations:", strings.Join(getFileIntegrations(), ", "))
			os.Exit(1)
		}

		fmt.Printf("Changing paths for %s...\n", integration)

		result, err := collector.ChangeFilePaths(integration, paths)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		fmt.Println("Old paths:")
		for _, p := range result.OldPaths {
			fmt.Printf("  - %s\n", p)
		}
		fmt.Println("New paths:")
		for _, p := range paths {
			fmt.Printf("  - %s\n", p)
		}
		time.Sleep(5 * time.Second)

		return nil
	},
}

var showPathsCmd = &cobra.Command{
	Use:     "show-paths <integration>",
	Short:   "Show the configured paths for a file-based integration",
	Args:    cobra.ExactArgs(1),
	PreRunE: requireInstalled,
	RunE: func(cmd *cobra.Command, args []string) error {
		integration := args[0]

		// Validate this is a file-type integration
		if config.ValidateModuleType(integration) != "file" {
			fmt.Printf("Error: %s is not a file-based integration\n", integration)
			fmt.Println("Valid file integrations:", strings.Join(getFileIntegrations(), ", "))
			os.Exit(1)
		}

		paths, err := collector.GetFileIntegrationPaths(integration)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		fmt.Printf("Configured paths for %s:\n", integration)
		for _, p := range paths {
			fmt.Printf("  - %s\n", p)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(changePathsCmd)
	rootCmd.AddCommand(showPathsCmd)
}
