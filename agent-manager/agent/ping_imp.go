package agent

import (
	"io"
	"log"
	"time"
)

func (s *Grpc) Ping(stream PingService_PingServer) error {
	for {
		req, err := stream.Recv()
		key := req.GetKey()
		if err == io.EOF {
			log.Printf("Client disconnected from Ping service: %s", key)
			break
		}
		if err != nil {
			return err
		}
		err = lastSeenService.Set(key, time.Now())
		if err != nil {
			log.Printf("unable to update last seen for: %s with error:%s", key, err)
		}
	}
	return nil
}
