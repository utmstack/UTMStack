package upstream

import (
	"context"
	"time"

	"github.com/utmstack/UTMStack/collectors/forwarder/config"
	"github.com/utmstack/UTMStack/collectors/forwarder/utils"
)

var (
	timeToSleep  = 10 * time.Second
	pingInterval = 15 * time.Second
)

func StartPing(cnf *config.Config, ctx context.Context) {
	var connErrLogged, streamErrLogged bool

	for {
		connection, err := GetAgentManagerConnection(cnf)
		if err != nil {
			LogConnectionError(err, "Agent Manager", &connErrLogged)
			time.Sleep(timeToSleep)
			continue
		}

		client := NewPingServiceClient(connection)
		stream, err := client.Ping(ctx)
		if err != nil {
			LogStreamError(err, "Ping Stream", &connErrLogged)
			time.Sleep(timeToSleep)
			continue
		}

		utils.Logger.LogF(100, "Ping Stream started")
		connErrLogged = false

		ticker := time.NewTicker(pingInterval)

	pingLoop:
		for range ticker.C {
			err := stream.Send(&PingRequest{Type: ConnectorType_COLLECTOR})
			if err != nil {
				action := HandleGRPCStreamError(err, "error sending Ping request", &streamErrLogged)
				if action == ActionReconnect {
					break pingLoop
				}
				continue
			}

			streamErrLogged = false
			utils.Logger.LogF(100, "Ping request sent")
		}

		ticker.Stop()
	}
}
