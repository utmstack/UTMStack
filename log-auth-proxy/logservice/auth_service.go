package logservice

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/log-auth-proxy/agent"
	"github.com/utmstack/UTMStack/log-auth-proxy/config"
	"github.com/utmstack/UTMStack/log-auth-proxy/panelservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const maxMessageSize = 1024 * 1024 * 1024

type LogAuthService struct {
	Mutex              *sync.Mutex
	AgentKeyCache      []string
	ConnectionKeyCache string
}

func NewLogAuthService() *LogAuthService {
	authService := &LogAuthService{
		Mutex:              &sync.Mutex{},
		AgentKeyCache:      make([]string, 0),
		ConnectionKeyCache: "",
	}

	authService.syncConnectionKey()
	authService.syncAgentsKeys()

	return authService
}

func (auth *LogAuthService) SyncAuth() {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			auth.syncAgentsKeys()
			auth.syncConnectionKey()
		}
	}
}

func (auth *LogAuthService) syncAgentsKeys() {
	serverAddress := os.Getenv(config.UTMAgentManagerHostEnv)
	if serverAddress == "" {
		log.Fatalf("Failed to get the SERVER_ADDRESS ")
	}

	conn, err := grpc.Dial(serverAddress, grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize)))
	if err != nil {
		log.Printf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	agentClient := agent.NewAgentServiceClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = metadata.AppendToOutgoingContext(ctx, "internal-key", os.Getenv(config.UTMSharedKeyEnv))

	response, err := agentClient.ListAgents(ctx, &agent.ListRequest{
		PageNumber:  1,
		PageSize:    100000,
		SearchQuery: "",
		SortBy:      "",
	})

	if err != nil {
		log.Printf("Error sync agent keys: %v", err)
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

func (auth *LogAuthService) syncConnectionKey() {
	panelKey, err := panelservice.GetConnectionKey()
	if err != nil {
		return
	}
	auth.Mutex.Lock()
	auth.ConnectionKeyCache = string(panelKey)
	auth.Mutex.Unlock()
}

func (auth *LogAuthService) IsAgentKeyValid(agentKey string) bool {
	for _, a := range auth.AgentKeyCache {
		if a == agentKey {
			return true
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
