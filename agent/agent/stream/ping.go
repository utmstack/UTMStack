package stream

import (
	"context"
	"io"
	"time"

	"github.com/quantfall/holmes"
	pb "github.com/utmstack/UTMStack/agent/agent/agent"
	"github.com/utmstack/UTMStack/agent/agent/configuration"
	"github.com/utmstack/UTMStack/agent/agent/utils"
)

// Call this function to start sending Ping requests
func StartPing(client pb.AgentServiceClient, ctx context.Context, cnf *configuration.Config, h *holmes.Logger) {
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
				h.Error("failed to start Ping Stream: %v", err)
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
			err := stream.Send(&pb.PingResponse{AgentKey: cnf.AgentKey, IsAlive: true})
			if err == io.EOF {
				// Server closed the stream
				h.Error("Server closed the stream: %v", err)
				break
			}
			if err != nil {
				h.Error("Error sending Ping request: %v", err)
				break
			}
		}
	}
}
