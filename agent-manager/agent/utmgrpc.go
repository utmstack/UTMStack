package agent

import (
	"fmt"
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
	UnimplementedCollectorConfigurationServiceServer
	UnimplementedPingServiceServer
	// Mutex to protect concurrent access to the agents map
	mu          sync.Mutex
	configMutex sync.Mutex

	// Map to store connected agents and their gRPC streams
	AgentStreamMap     map[string]*StreamAgent
	CollectorStreamMap map[string]*StreamCollector
	// AgentCache to store agentToken  and agentId
	cacheMutex          sync.Mutex
	cacheCollectorMutex sync.Mutex

	ResultChannel  map[string]chan *CommandResult
	resultChannelM sync.Mutex

	ConfigChannel  map[string]chan *ConfigKnowledge
	configChannelM sync.Mutex
}

type ResultCallback func(result *CommandResult)

func InitGrpc() (*Grpc, error) {
	gRPC := &Grpc{
		AgentStreamMap:     make(map[string]*StreamAgent),
		CollectorStreamMap: make(map[string]*StreamCollector),
		ResultChannel:      make(map[string]chan *CommandResult),
		ConfigChannel:      make(map[string]chan *ConfigKnowledge),
	}
	CacheAgent = make(map[uint]string)
	CacheCollector = make(map[uint]string)
	err := gRPC.LoadAgentCacheFromDatabase()
	err = gRPC.LoadCollectorsCacheFromDatabase()
	return gRPC, err
}

func (s *Grpc) authenticateAgent(stream AgentService_AgentStreamServer) (string, error) {
	agentMessage, err := stream.Recv()
	if err != nil {
		return "", fmt.Errorf("failed to receive auth response: %v", err)
	}
	authResponse := agentMessage.GetAuthResponse()
	if authResponse == nil {
		return "", fmt.Errorf("expected auth response, got: %v", agentMessage.GetAuthResponse())
	}
	return s.cacheAuthenticate(authResponse, ConnectorType_AGENT)
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
	return s.cacheAuthenticate(authResponse, ConnectorType_COLLECTOR)
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
		return &s.cacheMutex
	default:
		return &s.cacheMutex // or &s.cacheMutex as a fallback
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
