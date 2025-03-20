package agent

import (
	context "context"
	"fmt"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/conn"
	"github.com/utmstack/UTMStack/agent/models"
	"github.com/utmstack/UTMStack/agent/utils"
)

func UpdateAgent(cnf *config.Config, ctx context.Context) error {
	connection, err := conn.GetAgentManagerConnection(cnf)
	if err != nil {
		return fmt.Errorf("error connecting to Agent Manager: %v", err)
	}

	client := NewAgentServiceClient(connection)

	osInfo, err := utils.GetOsInfo()
	if err != nil {
		return fmt.Errorf("error getting os info: %v", err)
	}

	version := models.Version{}
	if utils.CheckIfPathExist(config.VersionPath) {
		err = utils.ReadJson(config.VersionPath, &version)
		if err != nil {
			utils.Logger.Fatal("error reading version file: %v", err)
		}
	} else {
		version.Version = "10.6.0"
	}

	request := &AgentRequest{
		Hostname:       osInfo.Hostname,
		Version:        version.Version,
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
