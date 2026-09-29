package agent

import (
	"context"
	"fmt"
	"strconv"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/models"
	"github.com/utmstack/UTMStack/agent/utils"
	"github.com/utmstack/UTMStack/shared/fs"
	"google.golang.org/grpc/metadata"
)

func RegisterAgent(cnf *config.Config, UTMKey string) error {
	connection, err := GetAgentManagerConnection(cnf)
	if err != nil {
		return fmt.Errorf("error connecting to Agent Manager: %v", err)
	}

	agentClient := NewAgentServiceClient(connection)
	ctx, cancel := context.WithCancel(context.Background())
	ctx = metadata.AppendToOutgoingContext(ctx, "connection-key", UTMKey)
	defer cancel()

	// A host that is being re-installed already has an agent record on the
	// manager. Present the credentials of that record so the manager can hand
	// the same key back; without this proof it refuses to re-issue the key of an
	// existing (hostname, mac), which is what stops one endpoint from claiming
	// another endpoint's command channel.
	if prev, err := config.GetCurrentConfig(); err == nil && prev.AgentID != 0 && prev.AgentKey != "" {
		ctx = metadata.AppendToOutgoingContext(ctx,
			"agent-id", strconv.FormatUint(uint64(prev.AgentID), 10),
			"agent-key", prev.AgentKey)
	}

	ip, err := utils.GetIPAddress()
	if err != nil {
		return fmt.Errorf("error getting ip address: %v", err)
	}

	osInfo, err := utils.GetOsInfo()
	if err != nil {
		return fmt.Errorf("error getting os info: %v", err)
	}

	version := models.Version{}
	err = fs.ReadJSON(config.VersionPath, &version)
	if err != nil {
		return fmt.Errorf("error reading version file: %v", err)
	}

	request := &AgentRequest{
		Ip:             ip,
		Hostname:       osInfo.Hostname,
		Os:             osInfo.OsType,
		Platform:       osInfo.Platform,
		Version:        version.Version,
		RegisterBy:     osInfo.CurrentUser,
		Mac:            osInfo.Mac,
		OsMajorVersion: osInfo.OsMajorVersion,
		OsMinorVersion: osInfo.OsMinorVersion,
		Aliases:        osInfo.Aliases,
		Addresses:      osInfo.Addresses,
	}

	response, err := agentClient.RegisterAgent(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to register agent: %v", err)
	}

	cnf.AgentID = uint(response.Id)
	cnf.AgentKey = response.Key

	utils.Logger.LogF(100, "Agent registered with ID: %v", cnf.AgentID)

	return nil
}
