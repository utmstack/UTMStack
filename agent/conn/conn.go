package conn

import (
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/utils"
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
	correlationConn      *grpc.ClientConn
	correlationConnOnce  sync.Once
	agentManagerConn     *grpc.ClientConn
	agentManagerConnOnce sync.Once
)

func GetAgentManagerConnection(cnf *config.Config) (*grpc.ClientConn, error) {
	var err error
	agentManagerConnOnce.Do(func() {
		agentManagerConn, err = connectToServer(cnf.Server, config.AgentManagerPort, cnf.SkipCertValidation)
		if err != nil {
			err = fmt.Errorf("error connecting to Agent Manager: %v", err)
		}
	})
	if err != nil {
		return nil, err
	}

	state := agentManagerConn.GetState()
	if state == connectivity.Shutdown || state == connectivity.TransientFailure {
		agentManagerConn.Close()
		agentManagerConn, err = connectToServer(cnf.Server, config.AgentManagerPort, cnf.SkipCertValidation)
		if err != nil {
			return nil, fmt.Errorf("error connecting to Agent Manager: %v", err)
		}
	}

	return agentManagerConn, nil
}

func GetCorrelationConnection(cnf *config.Config) (*grpc.ClientConn, error) {
	var err error
	correlationConnOnce.Do(func() {
		correlationConn, err = connectToServer(cnf.Server, config.LogAuthProxyPort, cnf.SkipCertValidation)
		if err != nil {
			err = fmt.Errorf("error connecting to Correlation: %v", err)
		}
	})
	if err != nil {
		return nil, err
	}

	state := correlationConn.GetState()
	if state == connectivity.Shutdown || state == connectivity.TransientFailure {
		correlationConn.Close()
		correlationConn, err = connectToServer(cnf.Server, config.LogAuthProxyPort, cnf.SkipCertValidation)
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
	var err error

	for {
		if connectionAttemps >= maxConnectionAttempts {
			return nil, fmt.Errorf("failed to connect to Server: %v", err)
		}

		conn, err = grpc.NewClient(
			serverAddress,
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize)),
			grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: skip})))
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
