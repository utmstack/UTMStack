package upstream

import (
	"context"
	"fmt"

	"github.com/utmstack/UTMStack/collectors/forwarder/config"
	"github.com/utmstack/UTMStack/collectors/forwarder/utils"
	"google.golang.org/grpc/metadata"
)

func RegisterCollector(cnf *config.Config, UTMKey string) error {
	connection, err := GetAgentManagerConnection(cnf)
	if err != nil {
		return fmt.Errorf("error connecting to Agent Manager: %v", err)
	}

	agentClient := NewAgentServiceClient(connection)
	ctx, cancel := context.WithCancel(context.Background())
	ctx = metadata.AppendToOutgoingContext(ctx, "connection-key", UTMKey)
	defer cancel()

	ip, err := utils.GetIPAddress()
	if err != nil {
		return fmt.Errorf("error getting ip address: %v", err)
	}

	osInfo, err := utils.GetOsInfo()
	if err != nil {
		return fmt.Errorf("error getting os info: %v", err)
	}

	request := &AgentRequest{
		Ip:             ip,
		Hostname:       osInfo.Hostname,
		Os:             osInfo.OsType,
		Platform:       osInfo.Platform,
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

	cnf.CollectorID = uint(response.Id)
	cnf.CollectorKey = response.Key

	utils.Logger.LogF(100, "Collector registered with ID: %v", cnf.CollectorID)

	return nil
}
