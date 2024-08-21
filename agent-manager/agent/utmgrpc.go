package agent

import (
	"context"
	"fmt"
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
	UnimplementedAgentGroupServiceServer
	UnimplementedConfigServiceServer

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

func (s *Grpc) InitPingSync() {
	lastSeenService.Start()
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

func getKeyFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Internal, "metadata is not provided")
	}
	key, ok := md["key"]
	if !ok || len(key) == 0 {
		return "", status.Error(codes.Unauthenticated, "authorization key is not provided")
	}
	return key[0], nil
}

func findAgentIdByKey(cache map[uint]string, key string) uint {
	for id, val := range cache {
		if val == key {
			return id
		}
	}
	return 0
}
