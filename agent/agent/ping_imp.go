package agent

import (
	"context"
	"time"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/utils"
)

var (
	timeToSleep  = 10 * time.Second
	pingInterval = 15 * time.Second
)

func StartPing(cnf *config.Config, ctx context.Context) {
	var connErrLogged, streamErrLogged bool

	for {
		if ctx.Err() != nil {
			return
		}

		connection, err := GetAgentManagerConnection(cnf)
		if err != nil {
			LogConnectionError(err, "Agent Manager", &connErrLogged)
			if !sleepCtx(ctx, timeToSleep) {
				return
			}
			continue
		}

		client := NewPingServiceClient(connection)
		stream, err := client.Ping(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			LogStreamError(err, "Ping Stream", &connErrLogged)
			if !sleepCtx(ctx, timeToSleep) {
				return
			}
			continue
		}

		utils.Logger.LogF(100, "Ping Stream started")
		connErrLogged = false

		ticker := time.NewTicker(pingInterval)

	pingLoop:
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				err := stream.Send(&PingRequest{Type: ConnectorType_AGENT})
				if err != nil {
					if ctx.Err() != nil {
						ticker.Stop()
						return
					}
					action := HandleGRPCStreamError(err, "error sending Ping request", &streamErrLogged)
					if action == ActionReconnect {
						break pingLoop
					}
					continue
				}

				streamErrLogged = false
				utils.Logger.LogF(100, "Ping request sent")
			}
		}

		ticker.Stop()
	}
}
