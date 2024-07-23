package agent

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/agent-manager/service"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	lastSeenService     = service.NewLastSeenService()
	agentService        = service.NewAgentService(lastSeenService)
	agentCommandService = service.NewAgentCommandService()
	//agentModuleService  = service.NewAgentModuleService()
	agentGroupService = service.NewAgentGroupService()
	collectorService  = service.NewCollectorService(lastSeenService)
	CacheAgentKey     map[uint]string
	CacheCollectorKey map[uint]string
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

	agentStreamMutex     sync.Mutex
	collectorStreamMutex sync.Mutex
	AgentStreamMap       map[string]AgentService_AgentStreamServer
	CollectorStreamMap   map[string]CollectorService_CollectorStreamServer

	cacheAgentKeyMutex     sync.Mutex
	cacheCollectorKeyMutex sync.Mutex

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
	CacheAgentKey = make(map[uint]string)
	CacheCollectorKey = make(map[uint]string)
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

func (s *Grpc) authenticateStream(stream interface{}) (string, error) {
	authResponse, err := s.GetStreamAuth(stream)
	if err != nil {
		return "", status.Error(codes.Unauthenticated, fmt.Sprintf("failed to get stream auth: %v", err))
	}

	id := uint(authResponse.Id)
	key, typ, err := convertTokenToKey(authResponse.Key)
	if err != nil {
		return "", status.Error(codes.Unauthenticated, fmt.Sprintf("failed to convert token to key: %v", err))
	}

	var cacheKey string
	var ok bool

	switch typ {
	case "Agent":
		s.cacheAgentKeyMutex.Lock()
		cacheKey, ok = CacheAgentKey[id]
		s.cacheAgentKeyMutex.Unlock()
	case "Collector":
		s.cacheCollectorKeyMutex.Lock()
		cacheKey, ok = CacheCollectorKey[id]
		s.cacheCollectorKeyMutex.Unlock()
	default:
		return "", status.Error(codes.Unauthenticated, "invalid key type")
	}

	if !ok || cacheKey != key {
		return "", status.Error(codes.Unauthenticated, "invalid key here")
	}

	return cacheKey, nil
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
			err := stream.Context().Err()
			if err == nil {
				time.Sleep(time.Second)
				return nil
			} else {
				time.Sleep(time.Second)
				attempts++
			}
		}
	}
	return fmt.Errorf("stream is not valid")
}

func findAgentIdByKey(cache map[uint]string, key string) uint {
	for id, val := range cache {
		if val == key {
			return id
		}
	}
	return 0
}

func ValidateKeyPairFromCache(token string, id uint) (string, string, bool) {
	key, typ, err := convertTokenToKey(token)
	if err != nil {
		return "", "", false
	}

	switch typ {
	case "Agent":
		for agentId, agentKey := range CacheAgentKey {
			if key == agentKey && id == agentId {
				return agentKey, "Agent", true
			}
		}
	case "Collector":
		for collId, collKey := range CacheCollectorKey {
			if key == collKey && id == collId {
				return collKey, "Collector", true
			}
		}
	}

	return "", "", false

}

func convertTokenToKey(token string) (string, string, error) {
	typ, key, found := strings.Cut(token, " ")
	if !found {
		return "", "", fmt.Errorf("invalid token type")
	}
	if typ == "Agent" || typ == "Collector" {
		return key, typ, nil
	}

	return "", "", fmt.Errorf("invalid token type")
}
