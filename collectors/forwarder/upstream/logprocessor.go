package upstream

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/collectors/forwarder/config"
	"github.com/utmstack/UTMStack/collectors/forwarder/utils"
)

// ProcessLogs reads from queue and streams logs directly to the backend.
// If the connection drops it reconnects and keeps going.
// Logs that can't be sent are dropped — no local buffering.
func ProcessLogs(cnf *config.Config, ctx context.Context, queue chan *plugins.Log) {
	var connErrLogged bool

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn, err := GetCorrelationConnection(cnf)
		if err != nil {
			if !connErrLogged {
				utils.Logger.ErrorF("error connecting to log-auth-proxy: %v", err)
				connErrLogged = true
			}
			time.Sleep(10 * time.Second)
			continue
		}

		client := plugins.NewIntegrationClient(conn)
		stream, err := client.ProcessLog(ctx)
		if err != nil {
			if !connErrLogged {
				utils.Logger.ErrorF("error creating log stream: %v", err)
				connErrLogged = true
			}
			time.Sleep(10 * time.Second)
			continue
		}

		connErrLogged = false
		utils.Logger.Info("Log stream connected")

		sendLogs(stream, ctx, queue)
	}
}

func sendLogs(stream plugins.Integration_ProcessLogClient, ctx context.Context, queue chan *plugins.Log) {
	for {
		select {
		case <-ctx.Done():
			return
		case log := <-queue:
			if err := stream.Send(log); err != nil {
				utils.Logger.ErrorF("error sending log, reconnecting: %v", err)
				return
			}
		}
	}
}
