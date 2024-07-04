package agent

import (
	context "context"
	"fmt"
	"os/user"
	"strconv"
	"time"

	"github.com/utmstack/UTMStack/agent/agent/config"
	"github.com/utmstack/UTMStack/agent/agent/utils"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	maxConnectionAttempts = 3
	initialReconnectDelay = 10 * time.Second
	maxReconnectDelay     = 60 * time.Second
)

func DeleteAgent(conn *grpc.ClientConn, cnf *config.Config) error {
	connectionAttemps := 0
	reconnectDelay := initialReconnectDelay

	agentClient := NewAgentServiceClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "key", cnf.AgentKey)
	ctx = metadata.AppendToOutgoingContext(ctx, "id", strconv.Itoa(int(cnf.AgentID)))

	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("error getting user: %v", err)
	}

	delet := &AgentDelete{
		DeletedBy: currentUser.Username,
	}

	for {
		if connectionAttemps >= maxConnectionAttempts {
			return fmt.Errorf("error removing UTMStack Agent from Agent Manager")
		}
		utils.Logger.Info("trying to remove UTMStack Agent from Agent Manager...")

		_, err = agentClient.DeleteAgent(ctx, delet)
		if err != nil {
			connectionAttemps++
			utils.Logger.Info("error removing UTMStack Agent from Agent Manager, trying again in %.0f seconds", reconnectDelay.Seconds())
			time.Sleep(reconnectDelay)
			reconnectDelay = utils.IncrementReconnectDelay(reconnectDelay, maxReconnectDelay)
			continue
		}

		break
	}

	utils.Logger.Info("UTMStack Agent removed successfully")
	return nil
}
