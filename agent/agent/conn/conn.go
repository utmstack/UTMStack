package conn

import (
	"fmt"
	"time"

	"github.com/quantfall/holmes"
	"github.com/utmstack/UTMStack/agent/agent/configuration"
	"github.com/utmstack/UTMStack/agent/agent/utils"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	maxMessageSize        = 1024 * 1024 * 1024
	maxConnectionAttempts = 3
	initialReconnectDelay = 10 * time.Second
	maxReconnectDelay     = 60 * time.Second
)

func ConnectToServer(cnf *configuration.Config, h *holmes.Logger, addrs, port string) (*grpc.ClientConn, error) {
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

		if cnf.SkipCertValidation {
			conn, err = grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize)))
			if err != nil {
				connectionAttemps++
				h.Info("error connecting to Server, trying again in %.0f seconds", reconnectDelay.Seconds())
				fmt.Printf("error connecting to Server, trying again in %.0f seconds\n", reconnectDelay.Seconds())
				time.Sleep(reconnectDelay)
				reconnectDelay = utils.IncrementReconnectDelay(reconnectDelay, maxReconnectDelay)
				continue
			}
		} else {
			tlsCredentials, err := utils.LoadTLSCredentials(configuration.GetCertPath())
			if err != nil {
				return nil, fmt.Errorf("failed to load TLS credentials: %v", err)
			}
			conn, err = grpc.Dial(serverAddress, grpc.WithTransportCredentials(tlsCredentials), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize)))
			if err != nil {
				connectionAttemps++
				h.Info("error connecting to Server, trying again in %.0f seconds", reconnectDelay.Seconds())
				fmt.Printf("error connecting to Server, trying again in %.0f seconds\n", reconnectDelay.Seconds())
				time.Sleep(reconnectDelay)
				reconnectDelay = utils.IncrementReconnectDelay(reconnectDelay, maxReconnectDelay)
				continue
			}
		}

		break
	}

	return conn, nil
}
