package agent

import (
	"context"
	"os/user"
	"strconv"

	"github.com/utmstack/UTMStack/utmstack-collector/config"
	"github.com/utmstack/UTMStack/utmstack-collector/conn"
	"github.com/utmstack/UTMStack/utmstack-collector/utils"
	"google.golang.org/grpc/metadata"
)

func DeleteAgent(cnf *config.Config) error {
	connection, err := conn.GetAgentManagerConnection(cnf)
	if err != nil {
		return utils.Logger.ErrorF("error connecting to Agent Manager: %v", err)
	}

	collectorClient := NewCollectorServiceClient(connection)
	ctx, cancel := context.WithCancel(context.Background())
	ctx = metadata.AppendToOutgoingContext(ctx, "key", cnf.CollectorKey)
	ctx = metadata.AppendToOutgoingContext(ctx, "id", strconv.Itoa(int(cnf.CollectorID)))
	ctx = metadata.AppendToOutgoingContext(ctx, "type", "collector")
	defer cancel()

	currentUser, err := user.Current()
	if err != nil {
		return utils.Logger.ErrorF("error getting user: %v", err)
	}

	delReq := &DeleteRequest{
		DeletedBy: currentUser.Username,
	}

	_, err = collectorClient.DeleteCollector(ctx, delReq)
	if err != nil {
		utils.Logger.ErrorF("error removing UTMStack Collector from Agent Manager %v", err)
	}

	utils.Logger.LogF(100, "UTMStack Collector removed successfully from agent manager")
	return nil
}
