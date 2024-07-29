package agent

import (
	"io"
	"time"

	"github.com/utmstack/UTMStack/agent-manager/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Grpc) Ping(stream PingService_PingServer) error {
	key, err := getKeyFromContext(stream.Context())
	if err != nil {
		return err
	}

	for {
		_, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		err = lastSeenService.Set(key, time.Now())
		if err != nil {
			utils.ALogger.ErrorF("unable to update last seen for: %s with error:%s", key, err)
		}
	}
}
