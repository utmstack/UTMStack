package agent

import (
	"context"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent/agent/config"
	"github.com/utmstack/UTMStack/agent/agent/conn"
	"github.com/utmstack/UTMStack/agent/agent/utils"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

var (
	timeToSleep  = time.Duration(10 * time.Second)
	pingInterval = time.Duration(15 * time.Second)
)

func StartPing(cnf *config.Config, ctx context.Context) {
	var connErrMsgWritten, errorLogged bool

	for {
		conn, err := conn.GetAgentManagerConnection(cnf)
		if err != nil {
			if !connErrMsgWritten {
				utils.Logger.ErrorF("error connecting to Agent Manager: %v", err)
				connErrMsgWritten = true
			}
			time.Sleep(timeToSleep)
			continue
		}

		client := NewPingServiceClient(conn)
		stream, err := client.Ping(ctx)
		if err != nil {
			if !connErrMsgWritten {
				utils.Logger.ErrorF("failed to start Ping Stream: %v", err)
				connErrMsgWritten = true
			}
			time.Sleep(timeToSleep)
			continue
		}

		connErrMsgWritten = false

		ticker := time.NewTicker(pingInterval)
		defer ticker.Stop()

		for range ticker.C {
			err := stream.Send(&PingRequest{Type: ConnectorType_AGENT})
			if err != nil {
				if strings.Contains(err.Error(), "EOF") {
					time.Sleep(timeToSleep)
					break
				}
				st, ok := status.FromError(err)
				if ok && (st.Code() == codes.Unavailable || st.Code() == codes.Canceled) {
					if !errorLogged {
						utils.Logger.ErrorF("error sending Ping request: %v", err)
						errorLogged = true
					}
					time.Sleep(timeToSleep)
					break
				} else {
					if !errorLogged {
						utils.Logger.ErrorF("error sending Ping request: %v", err)
						errorLogged = true
					}
					time.Sleep(timeToSleep)
					continue
				}
			}
			errorLogged = false
		}
	}
}
