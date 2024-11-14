package conn

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/utmstack/UTMStack/agent/service/config"
	"github.com/utmstack/UTMStack/agent/service/utils"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
)

const (
	maxMessageSize        = 1024 * 1024 * 1024
	maxConnectionAttempts = 3
	initialReconnectDelay = 10 * time.Second
	maxReconnectDelay     = 60 * time.Second
)

var (
	correlationConn     *grpc.ClientConn
	agentManagerConn    *grpc.ClientConn
	isConnectedFistTime bool
)

func EstablishConnectionsFistTime(cnf *config.Config) error {
	var err error
	agentManagerConn, err = connectToServer(cnf.Server, config.AGENTMANAGERPORT, cnf.SkipCertValidation)
	if err != nil {
		return fmt.Errorf("error connecting to Agent Manager: %v", err)
	}

	correlationConn, err = connectToServer(cnf.Server, config.CORRELATIONLOGSPORT, cnf.SkipCertValidation)
	if err != nil {
		return fmt.Errorf("error connecting to Correlation: %v", err)
	}

	isConnectedFistTime = true
	utils.Logger.Info("Server connections established successfully")
	return nil
}

func GetAgentManagerConnection(cnf *config.Config) (*grpc.ClientConn, error) {
	if !isConnectedFistTime {
		for {
			if isConnectedFistTime {
				return agentManagerConn, nil
			}
			time.Sleep(5 * time.Second)
		}
	}

	state := agentManagerConn.GetState()
	if state == connectivity.Shutdown || state == connectivity.TransientFailure {
		agentManagerConn.Close()
		var err error
		agentManagerConn, err = connectToServer(cnf.Server, config.AGENTMANAGERPORT, cnf.SkipCertValidation)
		if err != nil {
			return nil, fmt.Errorf("error connecting to Agent Manager: %v", err)
		}
	}

	return agentManagerConn, nil
}

func GetCorrelationConnection(cnf *config.Config) (*grpc.ClientConn, error) {
	if !isConnectedFistTime {
		for {
			if isConnectedFistTime {
				return correlationConn, nil
			}
			time.Sleep(5 * time.Second)
		}
	}

	state := correlationConn.GetState()
	if state == connectivity.Shutdown || state == connectivity.TransientFailure {
		correlationConn.Close()

		var err error
		correlationConn, err = connectToServer(cnf.Server, config.CORRELATIONLOGSPORT, cnf.SkipCertValidation)
		if err != nil {
			return nil, fmt.Errorf("error connecting to Correlation: %v", err)
		}
	}

	return correlationConn, nil
}

func connectToServer(addrs, port string, skip bool) (*grpc.ClientConn, error) {
	connectionAttemps := 0
	reconnectDelay := initialReconnectDelay

	serverAddress := addrs + ":" + port
	var conn *grpc.ClientConn

	for {
		if connectionAttemps >= maxConnectionAttempts {
			return nil, fmt.Errorf("failed to connect to Server")
		}

		var tlsCredentials credentials.TransportCredentials
		var err error

		if skip {
			tlsConfig := &tls.Config{
				InsecureSkipVerify: true,
			}
			tlsCredentials = credentials.NewTLS(tlsConfig)
		} else {
			tlsCredentials, err = utils.LoadTLSCredentials(config.GetCertPath())
			if err != nil {
				return nil, fmt.Errorf("failed to load TLS credentials: %v", err)
			}
		}

		conn, err = grpc.NewClient(serverAddress, grpc.WithTransportCredentials(tlsCredentials), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize)))
		if err != nil {
			connectionAttemps++
			utils.Logger.ErrorF("error connecting to Server, trying again in %.0f seconds", reconnectDelay.Seconds())
			time.Sleep(reconnectDelay)
			reconnectDelay = utils.IncrementReconnectDelay(reconnectDelay, maxReconnectDelay)
			continue
		}

		break
	}

	return conn, nil
}
