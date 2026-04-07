package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/utmstack/UTMStack/agent/agent"
)

var changeRetentionCmd = &cobra.Command{
	Use:     "change-retention <new_retention>",
	Short:   "Change the log retention",
	Long:    "Change the log retention to <new_retention>. Retention must be a number of megabytes. Example: 20",
	Args:    cobra.ExactArgs(1),
	PreRunE: requireInstalled,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Changing log retention ...")
		retention := args[0]

		if err := agent.SetDataRetention(retention); err != nil {
			fmt.Println("Error trying to change retention: ", err)
			os.Exit(1)
		}

		fmt.Printf("Retention changed correctly to %s\n", retention)
		time.Sleep(5 * time.Second)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(changeRetentionCmd)
}
