package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	edrconfig "github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/shared/fs"
	"github.com/utmstack/UTMStack/shared/svc"
)

var edrStatusCmd = &cobra.Command{
	Use:     "edr-status",
	Short:   "Show UTMStack EDR module status",
	PreRunE: requireInstalled,
	RunE: func(cmd *cobra.Command, args []string) error {
		st, _ := svc.Status(edrconfig.ServiceName)
		fmt.Printf("Service: %s\n", st)
		if fs.Exists(edrconfig.StatusFile) {
			lines, _ := fs.ReadLines(edrconfig.StatusFile)
			for _, l := range lines {
				fmt.Println(l)
			}
		}
		return nil
	},
}

func init() { rootCmd.AddCommand(edrStatusCmd) }
