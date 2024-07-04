package agent

import (
	context "context"
	"fmt"
	"strings"

	"github.com/utmstack/UTMStack/agent/agent/config"
	"github.com/utmstack/UTMStack/agent/agent/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func RegisterAgent(conn *grpc.ClientConn, cnf *config.Config, UTMKey string) error {
	agentClient := NewAgentServiceClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
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
		Version:        osInfo.Version,
		RegisterBy:     osInfo.CurrentUser,
		Mac:            osInfo.Mac,
		OsMajorVersion: osInfo.OsMajorVersion,
		OsMinorVersion: osInfo.OsMinorVersion,
		Aliases:        osInfo.Aliases,
		Addresses:      osInfo.Addresses,
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "connection-key", UTMKey)
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
