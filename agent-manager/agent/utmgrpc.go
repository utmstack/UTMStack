package agent

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/agent-manager/service"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	lastSeenService     = service.NewLastSeenService()
	agentService        = service.NewAgentService(lastSeenService)
	agentCommandService = service.NewAgentCommandService()
	//agentModuleService  = service.NewAgentModuleService()
	agentGroupService = service.NewAgentGroupService()
	collectorService  = service.NewCollectorService(lastSeenService)
	CacheAgent        map[uint]string
	CacheCollector    map[uint]string
)

type Grpc struct {
	UnimplementedAgentServiceServer
	UnimplementedPanelServiceServer
	UnimplementedAgentConfigServiceServer
	UnimplementedAgentMalwareServiceServer
	UnimplementedAgentGroupServiceServer

	UnimplementedCollectorServiceServer
	UnimplementedPanelCollectorServiceServer
	UnimplementedPingServiceServer
	// Mutex to protect concurrent access to the agents map
	agentStreamMutex     sync.Mutex
	collectorStreamMutex sync.Mutex

	// Map to store connected agents and their gRPC streams
	AgentStreamMap     map[string]AgentService_AgentStreamServer
	CollectorStreamMap map[string]CollectorService_CollectorStreamServer
	// AgentCache to store agentToken  and agentId
	cacheAgentMutex     sync.Mutex
	cacheCollectorMutex sync.Mutex

	ResultChannel  map[string]chan *CommandResult
	resultChannelM sync.Mutex

	PendingConfigs map[string]string
	pendingConfigM sync.Mutex
}

type ResultCallback func(result *CommandResult)

func InitGrpc() (*Grpc, error) {
	gRPC := &Grpc{
		AgentStreamMap:     make(map[string]AgentService_AgentStreamServer),
		CollectorStreamMap: make(map[string]CollectorService_CollectorStreamServer),
		ResultChannel:      make(map[string]chan *CommandResult),
		PendingConfigs:     make(map[string]string),
	}
	CacheAgent = make(map[uint]string)
	CacheCollector = make(map[uint]string)
	err := gRPC.LoadAgentCacheFromDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to load agent cache from database: %v", err)
	}
	err = gRPC.LoadCollectorsCacheFromDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to load collector cache from database: %v", err)
	}
	return gRPC, err
}

func (s *Grpc) authenticateConnector(stream interface{}, connectorType ConnectorType) (string, error) {
	authResponse, err := s.GetStreamAuth(stream)
	if err != nil {
		return "", err
	}

	return s.cacheAuthenticate(&authResponse, connectorType)
}

func (s *Grpc) cacheAuthenticate(authResponse *AuthResponse, connectorType ConnectorType) (string, error) {
	key := authResponse.Key
	id := uint(authResponse.Id)
	mutex := s.getAuthCacheMutex(connectorType)
	cache := s.getAuthCache(connectorType)

	mutex.Lock()
	cachedToken, ok := cache[id]
	mutex.Unlock()

	if !ok || cachedToken != key {
		return "", fmt.Errorf("invalid key: %s", key)
	}

	return key, nil
}

func findAgentIdByKey(cache map[uint]string, key string) uint {
	for id, val := range cache {
		if val == key {
			return id
		}
	}
	return 0 // key not found
}

func (s *Grpc) getAuthCacheMutex(connectorType ConnectorType) *sync.Mutex {
	switch connectorType {
	case ConnectorType_COLLECTOR:
		return &s.cacheCollectorMutex
	case ConnectorType_AGENT:
		return &s.cacheAgentMutex
	default:
		return &s.cacheAgentMutex // or &s.cacheMutex as a fallback
	}
}

func (s *Grpc) getAuthCache(connectorType ConnectorType) map[uint]string {
	switch connectorType {
	case ConnectorType_COLLECTOR:
		return CacheCollector
	case ConnectorType_AGENT:
		return CacheAgent
	default:
		return CacheAgent // or &s.cacheMutex as a fallback
	}
}

func (s *Grpc) InitPingSync() {
	lastSeenService.Start()
}

func (s *Grpc) GetStreamAuth(stream interface{}) (AuthResponse, error) {
	var ctx context.Context
	switch s := stream.(type) {
	case AgentService_AgentStreamServer:
		ctx = s.Context()
	case CollectorService_CollectorStreamServer:
		ctx = s.Context()
	case PingService_PingServer:
		ctx = s.Context()
	default:
		return AuthResponse{}, fmt.Errorf("invalid stream type")
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return AuthResponse{}, fmt.Errorf("failed to get metadata from context")
	}

	ids, ok := md["id"]
	if !ok || len(ids) != 1 {
		return AuthResponse{}, fmt.Errorf("no id found in metadata")
	}
	idString := ids[0]
	id, err := strconv.ParseUint(idString, 10, 32)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("failed to convert id to uint32: %v", err)
	}

	keys, ok := md["key"]
	if !ok || len(keys) != 1 {
		return AuthResponse{}, fmt.Errorf("no key found in metadata")
	}
	key := keys[0]

	return AuthResponse{
		Id:  uint32(id),
		Key: key,
	}, nil
}

// Wait for the client to reconnect
func (s *Grpc) waitForReconnect(ctx context.Context, key string, connectorType ConnectorType) error {
	var stream grpc.ServerStream
	if connectorType == ConnectorType_COLLECTOR {
		stream = s.CollectorStreamMap[key]
	} else {
		stream = s.AgentStreamMap[key]
	}

	attempts := 0
	for attempts < 10 {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context canceled: %v", ctx.Err())
		default:
			// Check if the stream is still valid
			err := stream.Context().Err()
			if err == nil {
				// StreamAgent is still valid, return nil
				time.Sleep(time.Second)
				return nil
			} else {
				// Stream is not valid, wait for a moment and try again
				time.Sleep(time.Second)
				attempts++
			}
		}
	}
	return fmt.Errorf("stream is not valid")
}
