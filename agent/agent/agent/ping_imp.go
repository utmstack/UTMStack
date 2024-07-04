package agent

import (
	"context"
	"io"
	"time"

	"github.com/utmstack/UTMStack/agent/agent/config"
	"github.com/utmstack/UTMStack/agent/agent/utils"
)

// Call this function to start sending Ping requests
func StartPing(client PingServiceClient, ctx context.Context, cnf *config.Config) {
	connectionTime := 0 * time.Second
	reconnectDelay := config.InitialReconnectDelay
	var connErrMsgWritten bool

	for {
		if connectionTime >= config.MaxConnectionTime {
			connectionTime = 0 * time.Second
			reconnectDelay = config.InitialReconnectDelay
			continue
		}

		stream, err := client.Ping(ctx)
		if err != nil {
			if !connErrMsgWritten {
				utils.Logger.ErrorF("failed to start Ping Stream: %v", err)
				connErrMsgWritten = true
			}
			time.Sleep(reconnectDelay)
			connectionTime = utils.IncrementReconnectTime(connectionTime, reconnectDelay, config.MaxConnectionTime)
			reconnectDelay = utils.IncrementReconnectDelay(reconnectDelay, config.MaxReconnectDelay)
			continue
		}

		connErrMsgWritten = false

		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			err := stream.Send(&PingRequest{Type: ConnectorType_AGENT})
			if err == io.EOF {
				utils.Logger.ErrorF("server closed the stream: %v", err)
				break
			}
			if err != nil {
				utils.Logger.ErrorF("error sending Ping request: %v", err)
				break
			}
		}
	}
}
