package agent

import (
	"io"
	"time"

	"github.com/utmstack/UTMStack/agent-manager/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Grpc) Ping(stream PingService_PingServer) error {
	authResponse, err := s.GetStreamAuth(stream)
	if err != nil {
		return err
	}

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		key, err := s.cacheAuthenticate(&authResponse, req.Type)
		if err != nil || key == "" {
			return status.Error(codes.Unauthenticated, "authorization key is not provided or is invalid")
		}
		err = lastSeenService.Set(key, time.Now())
		if err != nil {
			util.Logger.ErrorF("unable to update last seen for: %s with error:%s", key, err)
		}
	}
}
