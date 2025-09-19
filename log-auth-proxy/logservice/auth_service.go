package logservice

import (
	"context"
	"crypto/tls"
	"os"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/log-auth-proxy/agent"
	"github.com/utmstack/UTMStack/log-auth-proxy/config"
	"github.com/utmstack/UTMStack/log-auth-proxy/panelservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

const maxMessageSize = 1024 * 1024 * 1024

type LogAuthService struct {
	Mutex              *sync.Mutex
	CollectorKeyCache  []string
	AgentKeyCache      []string
	ConnectionKeyCache string
}

func NewLogAuthService() *LogAuthService {
	authService := &LogAuthService{
		Mutex:              &sync.Mutex{},
		CollectorKeyCache:  make([]string, 0),
		AgentKeyCache:      make([]string, 0),
		ConnectionKeyCache: "",
	}

	authService.syncConnectionKey()
	authService.syncKeys(agent.ConnectorType_AGENT)
	authService.syncKeys(agent.ConnectorType_COLLECTOR)

	return authService
}

func (auth *LogAuthService) SyncAuth() {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		auth.syncKeys(agent.ConnectorType_COLLECTOR)
		auth.syncKeys(agent.ConnectorType_AGENT)
		auth.syncConnectionKey()
	}
}

func (auth *LogAuthService) syncKeys(typ agent.ConnectorType) {
	serverAddress := os.Getenv(config.UTMAgentManagerHostEnv)
	if serverAddress == "" {
		catcher.Error("Failed to get the SERVER_ADDRESS", nil, map[string]any{})
		os.Exit(1)
	}

	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	creds := credentials.NewTLS(tlsConfig)
	opts := grpc.WithTransportCredentials(creds)

	conn, err := grpc.NewClient(serverAddress, opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize)))
	if err != nil {
		catcher.Error("Failed to connect to gRPC server", err, map[string]any{})
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "internal-key", os.Getenv(config.UTMSharedKeyEnv))

	switch typ {
	case agent.ConnectorType_COLLECTOR:
		collectorClient := agent.NewCollectorServiceClient(conn)
		response, err := collectorClient.ListCollector(ctx, &agent.ListRequest{
			PageNumber:  1,
			PageSize:    100000,
			SearchQuery: "",
			SortBy:      "",
		})
		if err != nil {
			catcher.Error("Error sync collector keys", err, map[string]any{})
			return
		}

		collectorKeys := make([]string, 0, len(response.Rows))
		for _, row := range response.Rows {
			collectorKeys = append(collectorKeys, row.CollectorKey)
		}

		auth.Mutex.Lock()
		auth.CollectorKeyCache = collectorKeys
		auth.Mutex.Unlock()

	case agent.ConnectorType_AGENT:
		agentClient := agent.NewAgentServiceClient(conn)
		response, err := agentClient.ListAgents(ctx, &agent.ListRequest{
			PageNumber:  1,
			PageSize:    100000,
			SearchQuery: "",
			SortBy:      "",
		})
		if err != nil {
			catcher.Error("Error sync agent keys", err, map[string]any{})
			return
		}

		agentKeys := make([]string, 0, len(response.Rows))
		for _, row := range response.Rows {
			agentKeys = append(agentKeys, row.AgentKey)
		}
		auth.Mutex.Lock()
		auth.AgentKeyCache = agentKeys
		auth.Mutex.Unlock()
	}
}

func (auth *LogAuthService) syncConnectionKey() {
	panelKey, err := panelservice.GetConnectionKey()
	if err != nil {
		catcher.Error("Failed to get connection key", err, map[string]any{})
		return
	}
	auth.Mutex.Lock()
	auth.ConnectionKeyCache = string(panelKey)
	auth.Mutex.Unlock()
}

func (auth *LogAuthService) IsKeyValid(key string, typ agent.ConnectorType) bool {
	switch typ {
	case agent.ConnectorType_AGENT:
		for _, a := range auth.AgentKeyCache {
			if a == key {
				return true
			}
		}
	case agent.ConnectorType_COLLECTOR:
		for _, c := range auth.CollectorKeyCache {
			if c == key {
				return true
			}
		}
	}

	return false
}

func (auth *LogAuthService) IsConnectionKeyValid(connectionKey string) bool {
	return auth.ConnectionKeyCache == connectionKey
}

func (auth *LogAuthService) GetConnectionKey() string {
	return auth.ConnectionKeyCache
}
