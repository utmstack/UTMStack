package agent

import (
	"fmt"
	"time"

	"github.com/utmstack/UTMStack/collector-installer/utils"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	maxMessageSize        = 1024 * 1024 * 1024
	maxConnectionAttempts = 3
	initialReconnectDelay = 10 * time.Second
	maxReconnectDelay     = 60 * time.Second
)

func ConnectToServer(h *utils.BeautyLogger, addrs, port string) (*grpc.ClientConn, error) {
	connectionAttemps := 0
	reconnectDelay := initialReconnectDelay

	// Connect to the gRPC server
	serverAddress := addrs + ":" + port
	var conn *grpc.ClientConn
	var err error

	for {
		if connectionAttemps >= maxConnectionAttempts {
			return nil, fmt.Errorf("failed to connect to Server")
		}

		h.WriteSimpleMessage("trying to connect to Server...")

		conn, err = grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize)))
		if err != nil {
			connectionAttemps++
			h.WriteSimpleMessage(fmt.Sprintf("error connecting to Server, trying again in %.0f seconds", reconnectDelay.Seconds()))
			time.Sleep(reconnectDelay)
			reconnectDelay = utils.IncrementReconnectDelay(reconnectDelay, maxReconnectDelay)
			continue
		}

		break
	}

	return conn, nil
}
