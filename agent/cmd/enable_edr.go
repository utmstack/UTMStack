package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/utmstack/UTMStack/agent/config"
	edrconfig "github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/shared/svc"
)

var enableEDRCmd = &cobra.Command{
	Use:     "enable-edr",
	Short:   "Enable and start the UTMStack EDR module",
	PreRunE: requireInstalled,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := edrconfig.Load()
		if err != nil {
			return err
		}
		if agentCnf, err := config.GetCurrentConfig(); err == nil {
			c.Server = agentCnf.Server
			c.SkipCertValidate = agentCnf.SkipCertValidation
		}
		c.Enabled = true
		if err := edrconfig.Save(c); err != nil {
			return err
		}
		if err := svc.Start(edrconfig.ServiceName); err != nil {
			return fmt.Errorf("starting EDR service: %w", err)
		}
		fmt.Println("UTMStack EDR enabled")
		return nil
	},
}

func init() { rootCmd.AddCommand(enableEDRCmd) }
