package logservice

import (
	"context"
	"strings"

	"github.com/utmstack/UTMStack/log-auth-proxy/config"
)

func (s *Grpc) ProcessLogs(_ context.Context, req *LogMessage) (*ReceivedMessage, error) {
	go s.OutputService.SendBulkLog(config.LogType(strings.ToLower(req.GetLogType())), req.Data)
	return &ReceivedMessage{
		Received: true,
		Message:  "log received",
	}, nil
}
