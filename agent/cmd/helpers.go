package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/utmstack/UTMStack/agent/utils"
)

func requireInstalled(cmd *cobra.Command, args []string) error {
	isInstalled, err := utils.CheckIfServiceIsInstalled("UTMStackAgent")
	if err != nil {
		return fmt.Errorf("error checking if service is installed: %v", err)
	}
	if !isInstalled {
		return fmt.Errorf("UTMStackAgent service is not installed")
	}
	return nil
}

func requireNotInstalled(cmd *cobra.Command, args []string) error {
	isInstalled, err := utils.CheckIfServiceIsInstalled("UTMStackAgent")
	if err != nil {
		return fmt.Errorf("error checking if service is installed: %v", err)
	}
	if isInstalled {
		return fmt.Errorf("UTMStackAgent service is already installed")
	}
	return nil
}
