package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/utmstack/UTMStack/collectors/forwarder/config"
	"github.com/utmstack/UTMStack/collectors/forwarder/utils"
)

func requireInstalled(cmd *cobra.Command, args []string) error {
	isInstalled, err := utils.CheckIfServiceIsInstalled(config.ServiceName)
	if err != nil {
		return fmt.Errorf("error checking if service is installed: %v", err)
	}
	if !isInstalled {
		return fmt.Errorf("%s service is not installed", config.ServiceName)
	}
	return nil
}

func requireNotInstalled(cmd *cobra.Command, args []string) error {
	isInstalled, err := utils.CheckIfServiceIsInstalled(config.ServiceName)
	if err != nil {
		return fmt.Errorf("error checking if service is installed: %v", err)
	}
	if isInstalled {
		return fmt.Errorf("%s service is already installed", config.ServiceName)
	}
	return nil
}


