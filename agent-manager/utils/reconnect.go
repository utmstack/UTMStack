package utils

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
)

func WaitForReconnect(ctx context.Context, stream grpc.ServerStream) error {
	attempts := 0
	for attempts < 10 {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context canceled: %v", ctx.Err())
		default:
			err := stream.Context().Err()
			if err == nil {
				time.Sleep(time.Second)
				return nil
			} else {
				time.Sleep(time.Second)
				attempts++
			}
		}
	}
	return fmt.Errorf("stream is not valid")
}
