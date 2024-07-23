package agent

import (
	"io"
	"time"

	"github.com/utmstack/UTMStack/agent-manager/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Grpc) Ping(stream PingService_PingServer) error {
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		key, err := s.authenticateStream(stream)
		if err != nil || key == "" {
			return status.Error(codes.Unauthenticated, "authorization key is not provided or is invalid")
		}
		err = lastSeenService.Set(key, time.Now())
		if err != nil {
			utils.ALogger.ErrorF("unable to update last seen for: %s with error:%s", key, err)
		}
	}
}
