package agent

import (
	"context"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/conn"
	"github.com/utmstack/UTMStack/agent/utils"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

var (
	timeToSleep  = 10 * time.Second
	pingInterval = 15 * time.Second
)

func StartPing(cnf *config.Config, ctx context.Context) {
	var connErrMsgWritten, errorLogged bool

	for {
		connection, err := conn.GetAgentManagerConnection(cnf)
		if err != nil {
			if !connErrMsgWritten {
				utils.Logger.ErrorF("error connecting to Agent Manager: %v", err)
				connErrMsgWritten = true
			} else {
				utils.Logger.LogF(100, "error connecting to Agent Manager: %v", err)
			}
			time.Sleep(timeToSleep)
			continue
		}

		client := NewPingServiceClient(connection)
		stream, err := client.Ping(ctx)
		if err != nil {
			if !connErrMsgWritten {
				utils.Logger.ErrorF("failed to start Ping Stream: %v", err)
				connErrMsgWritten = true
			} else {
				utils.Logger.LogF(100, "failed to start Ping Stream: %v", err)
			}
			time.Sleep(timeToSleep)
			continue
		}

		utils.Logger.LogF(100, "Ping Stream started")
		connErrMsgWritten = false

		ticker := time.NewTicker(pingInterval)

		for range ticker.C {
			err := stream.Send(&PingRequest{Type: ConnectorType_AGENT})
			if err != nil {
				if strings.Contains(err.Error(), "EOF") {
					utils.Logger.LogF(100, "error sending Ping request: %v", err)
					time.Sleep(timeToSleep)
					break
				}
				st, ok := status.FromError(err)
				if ok && (st.Code() == codes.Unavailable || st.Code() == codes.Canceled) {
					if !errorLogged {
						utils.Logger.ErrorF("error sending Ping request: %v", err)
						errorLogged = true
					} else {
						utils.Logger.LogF(100, "error sending Ping request: %v", err)
					}
					time.Sleep(timeToSleep)
					break
				} else {
					if !errorLogged {
						utils.Logger.ErrorF("error sending Ping request: %v", err)
						errorLogged = true
					} else {
						utils.Logger.LogF(100, "error sending Ping request: %v", err)
					}
					time.Sleep(timeToSleep)
					continue
				}
			}

			errorLogged = false
			utils.Logger.LogF(100, "Ping request sent")
		}

		ticker.Stop()
	}
}
