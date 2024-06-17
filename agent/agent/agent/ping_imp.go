package agent

import (
	"context"
	"io"
	"time"

	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/agent/agent/configuration"
	"github.com/utmstack/UTMStack/agent/agent/utils"
)

// Call this function to start sending Ping requests
func StartPing(client PingServiceClient, ctx context.Context, cnf *configuration.Config, h *logger.Logger) {
	connectionTime := 0 * time.Second
	reconnectDelay := configuration.InitialReconnectDelay
	var connErrMsgWritten bool

	for {
		if connectionTime >= configuration.MaxConnectionTime {
			connectionTime = 0 * time.Second
			reconnectDelay = configuration.InitialReconnectDelay
			continue
		}

		stream, err := client.Ping(ctx)
		if err != nil {
			if !connErrMsgWritten {
				h.ErrorF("failed to start Ping Stream: %v", err)
				connErrMsgWritten = true
			}
			time.Sleep(reconnectDelay)
			connectionTime = utils.IncrementReconnectTime(connectionTime, reconnectDelay, configuration.MaxConnectionTime)
			reconnectDelay = utils.IncrementReconnectDelay(reconnectDelay, configuration.MaxReconnectDelay)
			continue
		}

		connErrMsgWritten = false

		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			err := stream.Send(&PingRequest{Type: ConnectorType_AGENT})
			if err == io.EOF {
				// Server closed the stream
				h.ErrorF("Server closed the stream: %v", err)
				break
			}
			if err != nil {
				h.ErrorF("Error sending Ping request: %v", err)
				break
			}
		}
	}
}
