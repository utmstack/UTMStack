package main

import (
	"context"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/helpers"
	"github.com/utmstack/UTMStack/log-auth-proxy/agent"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	pCfg, e := helpers.PluginCfg[PluginConfig]("com.utmstack")
	if e != nil {
		return
	}

	serverAddress := pCfg.AgentManager
	if serverAddress == "" {
		helpers.Logger().ErrorF("failed to get the SERVER_ADDRESS ")
		os.Exit(1)
	}

	conn, err := grpc.NewClient(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize)))
	if err != nil {
		helpers.Logger().ErrorF("failed to connect to gRPC server: %v", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "internal-key", pCfg.InternalKey)

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
			if !strings.Contains(err.Error(), "error reading server preface: http2: frame too large") {
				helpers.Logger().ErrorF("error sync collector keys: %v", err)
			}
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
			if !strings.Contains(err.Error(), "error reading server preface: http2: frame too large") {
				helpers.Logger().ErrorF("error sync agent keys: %v", err)
			}
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
	panelKey, e := GetConnectionKey()
	if e != nil {
		return
	}

	auth.Mutex.Lock()
	auth.ConnectionKeyCache = string(panelKey)
	auth.Mutex.Unlock()
}

func (auth *LogAuthService) IsKeyValid(key string) bool {
	t, k, f := strings.Cut(key, " ")
	if !f {
		return false
	}
	switch t {
	case "Agent":
		for _, a := range auth.AgentKeyCache {
			if a == k {
				return true
			}
		}
	case "Collector":
		for _, c := range auth.CollectorKeyCache {
			if c == k {
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
