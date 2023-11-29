package agent

import (
	context "context"
	"fmt"
	"os/user"
	"strconv"
	"time"

	"github.com/quantfall/holmes"
	"github.com/utmstack/UTMStack/agent/agent/configuration"
	"github.com/utmstack/UTMStack/agent/agent/utils"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	maxConnectionAttempts = 3
	initialReconnectDelay = 10 * time.Second
	maxReconnectDelay     = 60 * time.Second
)

func DeleteAgent(conn *grpc.ClientConn, cnf *configuration.Config, h *holmes.Logger) error {
	connectionAttemps := 0
	reconnectDelay := initialReconnectDelay

	// Create a client for AgentService
	agentClient := NewAgentServiceClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "agent-key", cnf.AgentKey)
	ctx = metadata.AppendToOutgoingContext(ctx, "agent-id", strconv.Itoa(int(cnf.AgentID)))

	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("error getting user: %v", err)
	}

	delet := &AgentDelete{
		AgentKey:  cnf.AgentKey,
		DeletedBy: currentUser.Username,
	}

	for {
		if connectionAttemps >= maxConnectionAttempts {
			return fmt.Errorf("error removing UTMStack Agent from Agent Manager")
		}
		h.Info("trying to remove UTMStack Agent from Agent Manager...")

		_, err = agentClient.DeleteAgent(ctx, delet)
		if err != nil {
			connectionAttemps++
			h.Info("error removing UTMStack Agent from Agent Manager, trying again in %.0f seconds", reconnectDelay.Seconds())
			fmt.Printf("error removing UTMStack Agent from Agent Manager, trying again in %.0f seconds\n", reconnectDelay.Seconds())
			time.Sleep(reconnectDelay)
			reconnectDelay = utils.IncrementReconnectDelay(reconnectDelay, maxReconnectDelay)
			continue
		}

		break
	}

	h.Info("UTMStack Agent removed successfully")
	return nil
}
