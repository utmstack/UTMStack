package agent

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"time"
)

func (s *Grpc) Ping(stream PingService_PingServer) error {
	for {
		req, err := stream.Recv()
		auth := req.GetAuth()
		key, err := s.cacheAuthenticate(auth, req.Type)
		if err != nil || key == "" {
			return status.Error(codes.Unauthenticated, "authorization key is not provided or is invalid")
		}
		if err == io.EOF {
			log.Printf("ping for type %s with id %d ends wiht error: %s", req.Type, auth.Id, err)
		}
		if err != nil {
			return err
		}
		err = lastSeenService.Set(key, time.Now())
		if err != nil {
			log.Printf("unable to update last seen for: %s with error:%s", key, err)
		}

	}
}
