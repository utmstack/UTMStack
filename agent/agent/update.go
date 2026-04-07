package agent

import (
	context "context"
	"time"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/models"
	"github.com/utmstack/UTMStack/agent/utils"
	"github.com/utmstack/UTMStack/shared/fs"
)

const updateInterval = 5 * time.Minute

func UpdateAgent(cnf *config.Config, ctx context.Context) {
	var errLogged bool

	for {
		err := updateAgentOnce(cnf, ctx)
		if err != nil {
			if !errLogged {
				utils.Logger.ErrorF("error updating agent: %v", err)
				errLogged = true
			}
		} else {
			errLogged = false
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(updateInterval):
		}
	}
}

func updateAgentOnce(cnf *config.Config, ctx context.Context) error {
	connection, err := GetAgentManagerConnection(cnf)
	if err != nil {
		return err
	}

	client := NewAgentServiceClient(connection)

	osInfo, err := utils.GetOsInfo()
	if err != nil {
		return err
	}

	version := models.Version{}
	if err = fs.ReadJSON(config.VersionPath, &version); err != nil {
		return err
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
	return err
}
