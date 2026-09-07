package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/utmstack/UTMStack/agent/config"
)

var setAutoUpdateCmd = &cobra.Command{
	Use:   "set-auto-update <on|off>",
	Short: "Pause or resume automatic agent updates",
	Long: "Pause or resume automatic agent updates. Paused freezes the agent at whatever " +
		"version it's currently running — the updater will keep checking but won't apply " +
		"what the server offers. This is not the same as pinning to an arbitrary version: " +
		"there is no way to fetch a specific past version to downgrade to, only to stay on " +
		"the one already installed. Takes effect the next time the updater service starts.",
	Args:    cobra.ExactArgs(1),
	PreRunE: requireInstalled,
	RunE: func(cmd *cobra.Command, args []string) error {
		var pause bool
		switch args[0] {
		case "on":
			pause = false
		case "off":
			pause = true
		default:
			fmt.Println("Argument must be \"on\" or \"off\"")
			os.Exit(1)
		}

		cnf, err := config.GetCurrentConfig()
		if err != nil {
			fmt.Println("Error reading current config: ", err)
			os.Exit(1)
		}

		cnf.PauseAutoUpdate = pause
		if err := config.SaveConfig(cnf); err != nil {
			fmt.Println("Error saving config: ", err)
			os.Exit(1)
		}

		if pause {
			fmt.Println("Auto-update paused. Restart the updater service for this to take effect.")
		} else {
			fmt.Println("Auto-update resumed. Restart the updater service for this to take effect.")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(setAutoUpdateCmd)
}
