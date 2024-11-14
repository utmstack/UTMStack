package agent

import (
	context "context"
	"fmt"
	"strings"

	"github.com/utmstack/UTMStack/agent/service/config"
	"github.com/utmstack/UTMStack/agent/service/conn"
	"github.com/utmstack/UTMStack/agent/service/utils"
	"google.golang.org/grpc/metadata"
)

func RegisterAgent(cnf *config.Config, UTMKey string) error {
	conn, err := conn.GetAgentManagerConnection(cnf)
	if err != nil {
		return fmt.Errorf("error connecting to Agent Manager: %v", err)
	}

	agentClient := NewAgentServiceClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	ctx = metadata.AppendToOutgoingContext(ctx, "connection-key", UTMKey)
	defer cancel()

	ip, err := utils.GetIPAddress()
	if err != nil {
		return err
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
		Version:        config.AgentVersion,
		RegisterBy:     osInfo.CurrentUser,
		Mac:            osInfo.Mac,
		OsMajorVersion: osInfo.OsMajorVersion,
		OsMinorVersion: osInfo.OsMinorVersion,
		Aliases:        osInfo.Aliases,
		Addresses:      osInfo.Addresses,
	}

	response, err := agentClient.RegisterAgent(ctx, request)
	if err != nil {
		if strings.Contains(err.Error(), "hostname has already been registered") {
			return fmt.Errorf("failed to register agent: hostname has already been registered")
		}
		return fmt.Errorf("failed to register agent: %v", err)
	}

	cnf.AgentID = uint(response.Id)
	cnf.AgentKey = response.Key

	utils.Logger.Info("successfully registered agent")

	return nil
}
