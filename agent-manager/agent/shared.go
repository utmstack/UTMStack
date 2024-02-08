package agent

import (
	"fmt"
	"sync"

	"github.com/utmstack/UTMStack/agent-manager/service"
)

var (
	agentService        = service.NewAgentService()
	agentCommandService = service.NewAgentCommandService()
	agentModuleService  = service.NewAgentModuleService()
	agentGroupService   = service.NewAgentGroupService()
)

type Stream struct {
	token  string
	stream AgentService_AgentStreamServer
}

type Grpc struct {
	UnimplementedAgentServiceServer
	UnimplementedPanelServiceServer
	UnimplementedAgentConfigServiceServer
	UnimplementedAgentMalwareServiceServer
	UnimplementedAgentGroupServiceServer

	// Mutex to protect concurrent access to the agents map
	mu sync.Mutex

	// Map to store connected agents and their gRPC streams
	AgentStreamMap map[string]*Stream
	// Cache to store agentToken  and agentId
	cacheMutex sync.Mutex

	ResultChannel  map[string]chan *CommandResult
	resultChannelM sync.Mutex
}

var Cache map[uint]string

var (
	SecretVariablesCache = make(map[string]string)
	cacheSecretLock      sync.RWMutex
)

type ResultCallback func(result *CommandResult)

func (s *Grpc) authenticateAgent(stream AgentService_AgentStreamServer) (string, error) {
	// Receive the agent's response containing the authentication token
	agentMessage, err := stream.Recv()
	if err != nil {
		return "", fmt.Errorf("failed to receive auth response: %v", err)
	}

	authResponse := agentMessage.GetAuthResponse()
	if authResponse == nil {
		return "", fmt.Errorf("expected auth response, got: %v", agentMessage.GetAuthResponse())
	}

	// Validate the token (e.g., by checking it against the cache)
	token := authResponse.AgentKey
	agentId := uint(authResponse.AgentId)

	s.cacheMutex.Lock()
	cachedToken, ok := Cache[agentId]
	s.cacheMutex.Unlock()

	if !ok || cachedToken != token {
		return "", fmt.Errorf("invalid token: %s", token)
	}

	// If the token is valid, return it
	return token, nil
}

func findAgentIdByToken(cache map[uint]string, token string) uint {
	for key, val := range cache {
		if val == token {
			return key
		}
	}
	return 0 // key not found
}
