package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	edrconfig "github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/shared/svc"
)

var disableEDRCmd = &cobra.Command{
	Use:     "disable-edr",
	Short:   "Disable and stop the UTMStack EDR module",
	PreRunE: requireInstalled,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := svc.Stop(edrconfig.ServiceName); err != nil {
			return fmt.Errorf("stopping EDR service: %w", err)
		}
		c, err := edrconfig.Load()
		if err == nil {
			c.Enabled = false
			_ = edrconfig.Save(c)
		}
		fmt.Println("UTMStack EDR disabled")
		return nil
	},
}

func init() { rootCmd.AddCommand(disableEDRCmd) }
