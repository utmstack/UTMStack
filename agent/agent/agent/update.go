package agent

import (
	context "context"
	"fmt"

	"github.com/utmstack/UTMStack/agent/agent/utils"
)

func UpdateAgent(client AgentServiceClient, ctx context.Context) error {
	osInfo, err := utils.GetOsInfo()
	if err != nil {
		return fmt.Errorf("error getting os info: %v", err)
	}

	version, err := GetVersion()
	if err != nil {
		return fmt.Errorf("error getting agent version: %v", err)
	}

	request := &AgentRequest{
		Hostname:       osInfo.Hostname,
		Version:        version,
		Mac:            osInfo.Mac,
		OsMajorVersion: osInfo.OsMajorVersion,
		OsMinorVersion: osInfo.OsMinorVersion,
		Aliases:        osInfo.Aliases,
		Addresses:      osInfo.Addresses,
	}

	_, err = client.UpdateAgent(ctx, request)
	if err != nil {
		return fmt.Errorf("error updating agent: %v", err)
	}

	return nil
}
