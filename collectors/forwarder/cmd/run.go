package cmd

import (
	"github.com/spf13/cobra"
	"github.com/utmstack/UTMStack/collectors/forwarder/serv"
)

var runCmd = &cobra.Command{
	Use:     "run",
	Short:   "Run the UTMStackForwarder service",
	Args:    cobra.NoArgs,
	PreRunE: requireInstalled,
	Run: func(cmd *cobra.Command, args []string) {
		serv.RunService()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
