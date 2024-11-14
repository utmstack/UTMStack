package agent

import (
	context "context"
	"fmt"
	"os/user"
	"strconv"

	"github.com/utmstack/UTMStack/agent/service/config"
	"github.com/utmstack/UTMStack/agent/service/conn"
	"github.com/utmstack/UTMStack/agent/service/utils"
	"google.golang.org/grpc/metadata"
)

func DeleteAgent(cnf *config.Config) error {
	conn, err := conn.GetAgentManagerConnection(cnf)
	if err != nil {
		return fmt.Errorf("error connecting to Agent Manager: %v", err)
	}

	agentClient := NewAgentServiceClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	ctx = metadata.AppendToOutgoingContext(ctx, "key", cnf.AgentKey)
	ctx = metadata.AppendToOutgoingContext(ctx, "id", strconv.Itoa(int(cnf.AgentID)))
	ctx = metadata.AppendToOutgoingContext(ctx, "type", "agent")
	defer cancel()

	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("error getting user: %v", err)
	}

	delet := &DeleteRequest{
		DeletedBy: currentUser.Username,
	}

	_, err = agentClient.DeleteAgent(ctx, delet)
	if err != nil {
		utils.Logger.ErrorF("error removing UTMStack Agent from Agent Manager %v", err)
	}

	utils.Logger.Info("UTMStack Agent removed successfully")
	return nil
}
