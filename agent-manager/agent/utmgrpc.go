package agent

import (
	"fmt"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"sync"

	"github.com/utmstack/UTMStack/agent-manager/service"
)

var (
	lastSeenService     = service.NewLastSeenService()
	agentService        = service.NewAgentService(lastSeenService)
	agentCommandService = service.NewAgentCommandService()
	agentModuleService  = service.NewAgentModuleService()
	agentGroupService   = service.NewAgentGroupService()
	collectorService    = service.NewCollectorService(lastSeenService)
	CacheAgent          map[uint]string
	CacheCollector      map[uint]string
)

type StreamAgent struct {
	key    string
	stream AgentService_AgentStreamServer
}

type StreamCollector struct {
	key    string
	stream CollectorService_CollectorStreamServer
}

type Grpc struct {
	UnimplementedAgentServiceServer
	UnimplementedPanelServiceServer
	UnimplementedAgentConfigServiceServer
	UnimplementedAgentMalwareServiceServer
	UnimplementedAgentGroupServiceServer

	UnimplementedCollectorServiceServer
	UnimplementedPingServiceServer
	// Mutex to protect concurrent access to the agents map
	mu sync.Mutex

	// Map to store connected agents and their gRPC streams
	AgentStreamMap     map[string]*StreamAgent
	CollectorStreamMap map[string]*StreamCollector
	// AgentCache to store agentToken  and agentId
	cacheMutex          sync.Mutex
	cacheCollectorMutex sync.Mutex

	ResultChannel  map[string]chan *CommandResult
	resultChannelM sync.Mutex
}

type ResultCallback func(result *CommandResult)

func (s *Grpc) authenticateAgent(stream AgentService_AgentStreamServer) (string, error) {
	agentMessage, err := stream.Recv()
	if err != nil {
		return "", fmt.Errorf("failed to receive auth response: %v", err)
	}
	authResponse := agentMessage.GetAuthResponse()
	if authResponse == nil {
		return "", fmt.Errorf("expected auth response, got: %v", agentMessage.GetAuthResponse())
	}
	return s.cacheAuthenticate(authResponse, CacheAgent, models.AgentClass)
}

func (s *Grpc) authenticateCollector(stream CollectorService_CollectorStreamServer) (string, error) {
	agentMessage, err := stream.Recv()
	if err != nil {
		return "", fmt.Errorf("failed to receive auth response: %v", err)
	}
	authResponse := agentMessage.GetAuthResponse()
	if authResponse == nil {
		return "", fmt.Errorf("expected auth response, got: %v", agentMessage.GetAuthResponse())
	}
	return s.cacheAuthenticate(authResponse, CacheCollector, models.CollectorClass)
}

func (s *Grpc) cacheAuthenticate(authResponse *AuthResponse, cache map[uint]string, class models.LogCollectorClass) (string, error) {
	key := authResponse.Key
	id := uint(authResponse.Id)
	mutex := s.getAuthCacheMutex(class)
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

func (s *Grpc) getAuthCacheMutex(class models.LogCollectorClass) *sync.Mutex {
	switch class {
	case models.CollectorClass:
		return &s.cacheCollectorMutex
	case models.AgentClass:
		return &s.cacheMutex
	default:
		return nil // or &s.cacheMutex as a fallback
	}
}

func (s *Grpc) InitPingSync() {
	lastSeenService.Start()
}
