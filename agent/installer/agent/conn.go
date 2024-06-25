package agent

import (
	"fmt"
	"time"

	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/agent/installer/configuration"
	"github.com/utmstack/UTMStack/agent/installer/utils"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	maxMessageSize        = 1024 * 1024 * 1024
	maxConnectionAttempts = 3
	initialReconnectDelay = 10 * time.Second
	maxReconnectDelay     = 60 * time.Second
)

func ConnectToServer(h *logger.Logger, addrs, port, skip string) (*grpc.ClientConn, error) {
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

		h.Info("trying to connect to Server...")

		if skip == "yes" {
			conn, err = grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize)))
			if err != nil {
				connectionAttemps++
				h.Info("error connecting to Server, trying again in %.0f seconds", reconnectDelay.Seconds())
				time.Sleep(reconnectDelay)
				reconnectDelay = utils.IncrementReconnectDelay(reconnectDelay, maxReconnectDelay)
				continue
			}
		} else {
			tlsCredentials, err := utils.LoadTLSCredentials(configuration.GetCertPath())
			if err != nil {
				return nil, fmt.Errorf("failed to load TLS credentials: %v", err)
			}
			transportCredentials := credentials.NewTLS(tlsCredentials)
			conn, err = grpc.Dial(serverAddress, grpc.WithTransportCredentials(transportCredentials), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize)))
			if err != nil {
				connectionAttemps++
				h.Info("error connecting to Server, trying again in %.0f seconds", reconnectDelay.Seconds())
				time.Sleep(reconnectDelay)
				reconnectDelay = utils.IncrementReconnectDelay(reconnectDelay, maxReconnectDelay)
				continue
			}
		}

		break
	}

	return conn, nil
}
