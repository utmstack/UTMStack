package main

import (
	"context"
	"crypto/tls"
	"os"
	"strings"
	"sync"
	"time"

	go_sdk "github.com/threatwinds/go-sdk"
	"github.com/utmstack/UTMStack/plugins/utmstack-inputs/agent"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

const maxMessageSize = 1024 * 1024 * 1024

type LogAuthService struct {
	Mutex              *sync.Mutex
	CollectorKeyCache  map[uint]string
	AgentKeyCache      map[uint]string
	ConnectionKeyCache string
}

func NewLogAuthService() *LogAuthService {
	authService := &LogAuthService{
		Mutex:              &sync.Mutex{},
		CollectorKeyCache:  make(map[uint]string),
		AgentKeyCache:      make(map[uint]string),
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
	pCfg, e := go_sdk.PluginCfg[PluginConfig]("com.utmstack")
	if e != nil {
		return
	}

	serverAddress := pCfg.AgentManager
	if serverAddress == "" {
		go_sdk.Logger().ErrorF("failed to get the SERVER_ADDRESS ")
		os.Exit(1)
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	tlsCredentials := credentials.NewTLS(tlsConfig)
	conn, err := grpc.NewClient(serverAddress, grpc.WithTransportCredentials(tlsCredentials), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize)))
	if err != nil {
		go_sdk.Logger().ErrorF("failed to connect to gRPC server: %v", err)
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
				go_sdk.Logger().ErrorF("error sync collector keys: %v", err)
			}
			return
		}

		collectorKeys := make(map[uint]string, len(response.Rows))
		for _, row := range response.Rows {
			collectorKeys[uint(row.Id)] = row.CollectorKey
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
				go_sdk.Logger().ErrorF("error sync agent keys: %v", err)
			}
			return
		}

		agentKeys := make(map[uint]string, len(response.Rows))
		for _, row := range response.Rows {
			agentKeys[uint(row.Id)] = row.AgentKey
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

func (auth *LogAuthService) IsKeyValid(key string, id uint, typ string) bool {
	switch typ {
	case "agent":
		for agentId, agentKey := range auth.AgentKeyCache {
			if key == agentKey && id == agentId {
				return true
			}
		}
	case "collector":
		for collId, collKey := range auth.CollectorKeyCache {
			if key == collKey && id == collId {
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
