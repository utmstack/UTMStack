package agent

import (
	"io"
	"log"
	"time"
)

func (s *Grpc) Ping(stream PingService_PingServer) error {
	for {
		req, err := stream.Recv()
		auth := req.GetAuth()
		key, err := s.cacheAuthenticate(auth, req.Type)
		if err != nil {
			return err
		}
		if err == io.EOF {
			err = stream.Send(&PingResponse{
				IsAlive: true,
				Key:     key,
				Type:    req.Type,
			})
			if err != nil {
				return err
			}
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
