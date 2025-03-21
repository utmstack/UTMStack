package agent

import (
	context "context"
	"fmt"
	"os/user"
	"strconv"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/conn"
	"github.com/utmstack/UTMStack/agent/utils"
	"google.golang.org/grpc/metadata"
)

func DeleteAgent(cnf *config.Config) error {
	connection, err := conn.GetAgentManagerConnection(cnf)
	if err != nil {
		return fmt.Errorf("error connecting to Agent Manager: %v", err)
	}

	agentClient := NewAgentServiceClient(connection)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "key", cnf.AgentKey)
	ctx = metadata.AppendToOutgoingContext(ctx, "id", strconv.Itoa(int(cnf.AgentID)))
	ctx = metadata.AppendToOutgoingContext(ctx, "type", "agent")

	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("error getting user: %v", err)
	}

	delet := &AgentDelete{
		DeletedBy: currentUser.Username,
	}

	_, err = agentClient.DeleteAgent(ctx, delet)
	if err != nil {
		utils.Logger.ErrorF("error removing UTMStack Agent from Agent Manager %v", err)
	}

	utils.Logger.LogF(100, "UTMStack Agent removed successfully from agent manager")
	return nil
}
