package agent

import (
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

const (
	maxMessageSize        = 1024 * 1024 * 1024
	maxConnectionAttempts = 3
	initialReconnectDelay = 10 * time.Second
	maxReconnectDelay     = 60 * time.Second
)

type connEntry struct {
	mu   sync.Mutex
	conn *grpc.ClientConn
}

func (e *connEntry) getOrCreate(cnf *config.Config, port string, name string) (*grpc.ClientConn, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.conn != nil {
		state := e.conn.GetState()
		if state != connectivity.Shutdown && state != connectivity.TransientFailure {
			return e.conn, nil
		}
		e.conn.Close()
	}

	conn, err := connectToServer(cnf.Server, port, cnf.SkipCertValidation)
	if err != nil {
		return nil, fmt.Errorf("error connecting to %s: %v", name, err)
	}
	e.conn = conn
	return e.conn, nil
}

var (
	agentManagerEntry connEntry
	correlationEntry  connEntry
)

func GetAgentManagerConnection(cnf *config.Config) (*grpc.ClientConn, error) {
	return agentManagerEntry.getOrCreate(cnf, config.AgentManagerPort, "Agent Manager")
}

func GetCorrelationConnection(cnf *config.Config) (*grpc.ClientConn, error) {
	return correlationEntry.getOrCreate(cnf, config.LogAuthProxyPort, "Correlation")
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

		// Keepalive pings (30s) travel over the shared HTTP/2 connection and
		// reset grpc_read_timeout on the nginx hops in front of the agent
		// manager, so an idle AgentStream is no longer torn down after 900s.
		// PermitWithoutStream matters: pings keep flowing while AgentStream is
		// down and being redialed, and on the PingService-only window.
		conn, err = grpc.NewClient(
			serverAddress,
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize)),
			grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: skip})),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                30 * time.Second,
				Timeout:             10 * time.Second,
				PermitWithoutStream: true,
			}))
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
