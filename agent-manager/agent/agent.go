package agent

import (
	"context"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent-manager/config"

	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Grpc) AgentStream(stream AgentService_AgentStreamServer) error {
	// Authenticate the agent and get the agent's token
	agentKey, err := s.authenticateAgent(stream)
	if err != nil {
		return err
	}

	// Add the agent to the agents map
	s.mu.Lock()
	s.AgentStreamMap[agentKey] = &Stream{token: agentKey, stream: stream}
	s.mu.Unlock()

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			// Remove the agent from the agents map if the connection is closed
			s.mu.Lock()
			delete(s.AgentStreamMap, agentKey)
			s.mu.Unlock()
			// Wait for the client to reconnect
			log.Printf("waiting for client to reconnect...")
			err = waitForReconnect(s.AgentStreamMap[agentKey].stream.Context(), s)
			if err != nil {
				log.Printf("failed to reconnect to client: %v", err)
			}

			// Reauthenticate the client and add it back to the agents map
			agentKey, err = s.authenticateAgent(stream)
			if err != nil {
				return err
			}
			s.mu.Lock()
			s.AgentStreamMap[agentKey] = &Stream{token: agentKey, stream: stream}
			s.mu.Unlock()
			return nil
		}
		if err != nil {
			return err
		}

		switch msg := in.StreamMessage.(type) {
		case *BidirectionalStream_Command:
			// Validate the internal key for incoming commands
			if msg.Command.GetInternalKey() != config.GetInternalKey() {
				log.Printf("unauthorized command attempt detected")
				continue // Skip processing this unauthorized command
			}

		case *BidirectionalStream_Result:
			// Handle the received command (replace with your logic)
			log.Printf("Received command: %s", msg.Result.CmdId)

			cmdID := msg.Result.GetCmdId()

			// Send the result back to the server
			if err := stream.Send(&BidirectionalStream{
				StreamMessage: &BidirectionalStream_Result{
					Result: &CommandResult{
						AgentKey:   msg.Result.AgentKey,
						Result:     msg.Result.Result,
						CmdId:      cmdID,
						ExecutedAt: msg.Result.ExecutedAt,
					},
				},
			}); err != nil {
				log.Printf("Failed to send result to server: %v", err)
			}
			s.resultChannelM.Lock()
			if resultChan, ok := s.ResultChannel[cmdID]; ok {
				resultChan <- &CommandResult{
					AgentKey:   msg.Result.AgentKey,
					Result:     msg.Result.Result,
					CmdId:      cmdID,
					ExecutedAt: msg.Result.ExecutedAt,
				}

			} else {
				log.Printf("Failed to find result channel for CmdID: %s", cmdID)
			}
			s.resultChannelM.Unlock()
		}
	}
}

func (s *Grpc) replaceSecretValues(input string) string {
	pattern := regexp.MustCompile(`\$\[(\w+):([^\]]+)\]`)
	return pattern.ReplaceAllStringFunc(input, func(match string) string {
		matches := pattern.FindStringSubmatch(match)
		if len(matches) < 3 {
			return match // In case of no match, return the original
		}
		encryptedValue := matches[2]
		// Decrypt and add to cache if not found
		decryptedValue, _ := util.DecryptValue(encryptedValue)
		addToSecretCache(match, decryptedValue)
		return decryptedValue
	})
}

func addToSecretCache(key, decryptedValue string) {
	cacheSecretLock.Lock()
	defer cacheSecretLock.Unlock()
	SecretVariablesCache[key] = decryptedValue
}

func hideSecretsInCommandResult(result string) string {
	cacheSecretLock.RLock()
	defer cacheSecretLock.RUnlock()
	for _, decryptedValue := range SecretVariablesCache {
		result = strings.ReplaceAll(result, decryptedValue, strings.Repeat("*", len(decryptedValue)))
	}
	return result
}

func (s *Grpc) ProcessCommand(stream PanelService_ProcessCommandServer) error {
	for {
		cmd, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// Get the agent from the agents map
		agentStream, ok := s.AgentStreamMap[cmd.AgentKey]
		if !ok {
			return status.Errorf(codes.NotFound, "agent not found or is disconnected")
		}

		if cmd.GetOriginId() == "" {
			return status.Errorf(codes.NotFound, "agent origin ID not provided")
		}

		if cmd.GetOriginType() == "" {
			return status.Errorf(codes.NotFound, "agent origin TYPE not provided")
		}

		if cmd.GetReason() == "" {
			return status.Errorf(codes.NotFound, "agent command reason not provided")
		}

		// Generate a unique command ID
		cmdID := uuid.New().String()

		// Create a channel for the agent result and store it in the map
		s.resultChannelM.Lock()
		resultChan := make(chan *CommandResult)
		s.ResultChannel[cmdID] = resultChan
		s.resultChannelM.Unlock()

		createHistoryCommand(cmd, cmdID)
		// Send the command to the agent along with the command ID
		err = agentStream.stream.Send(&BidirectionalStream{
			StreamMessage: &BidirectionalStream_Command{
				Command: &UtmCommand{
					AgentKey:    cmd.AgentKey,
					Command:     s.replaceSecretValues(cmd.Command),
					CmdId:       cmdID,
					InternalKey: config.GetInternalKey(),
				},
			},
		})
		if err != nil {
			return err
		}

		// Wait for the result from the agent
		result := <-resultChan
		result.Result = hideSecretsInCommandResult(result.Result)
		updateHistoryCommand(result, cmdID)
		// Send the result back to the PanelService
		err = stream.Send(result)
		if err != nil {
			return err
		}

		// Remove the result channel from the map
		s.resultChannelM.Lock()
		delete(s.ResultChannel, cmdID)
		s.resultChannelM.Unlock()
	}
}

func (s *Grpc) Ping(stream AgentService_PingServer) error {
	for {
		req, err := stream.Recv()
		agentKey := req.GetAgentKey()
		if err == io.EOF {
			log.Printf("Client disconnected from Ping service: %s", agentKey)
			break
		}
		if err != nil {
			return err
		}

		err = agentService.SetAgentLastSeen(agentKey)
		// Send a response to indicate that the agent is alive
		if err != nil {
			log.Printf("unable to update last seen for: %s with error:%s", agentKey, err)
		}
	}
	return nil
}

func (s *Grpc) RegisterAgent(ctx context.Context, req *AgentRequest) (*AgentResponse, error) {
	agent := &models.Agent{
		Ip:             req.GetIp(),
		Hostname:       req.GetHostname(),
		Os:             req.GetOs(),
		Platform:       req.GetPlatform(),
		Version:        req.GetVersion(),
		RegisterBy:     req.GetRegisterBy(),
		Mac:            req.GetMac(),
		OsMajorVersion: req.GetOsMajorVersion(),
		OsMinorVersion: req.GetOsMinorVersion(),
		Aliases:        req.GetAliases(),
		Addresses:      req.GetAddresses(),
	}

	oldAgent, err := s.GetAgentByHostname(ctx, &Hostname{Hostname: agent.Hostname})
	if err == nil {
		if oldAgent.Ip == agent.Ip {
			return &AgentResponse{
				Id:       oldAgent.Id,
				AgentKey: oldAgent.AgentKey,
			}, nil
		} else {
			return nil, status.Errorf(codes.AlreadyExists, "hostname has already been registered")
		}
	}

	token := uuid.New().String()
	agent.AgentKey = token
	err = agentService.Create(agent)
	if err != nil {
		return nil, err
	}
	// Update the cache
	s.cacheMutex.Lock()
	Cache[agent.ID] = token
	s.cacheMutex.Unlock()
	err = agentService.SetAgentLastSeen(token)
	if err != nil {
		return nil, err
	}
	res := &AgentResponse{
		Id:       uint32(agent.ID),
		AgentKey: token,
	}
	return res, nil
}

func (s *Grpc) DeleteAgent(ctx context.Context, req *AgentDelete) (*AgentResponse, error) {
	// ...
	// Delete the agent from the database and get its id
	id, err := agentService.Delete(uuid.MustParse(req.AgentKey), req.DeletedBy)
	if err != nil {
		return &AgentResponse{}, status.Error(codes.Internal, fmt.Sprintf("unable to delete agent: %v", err.Error()))
	}
	// Update the cache
	s.cacheMutex.Lock()
	delete(s.AgentStreamMap, req.AgentKey)
	s.cacheMutex.Unlock()
	return &AgentResponse{
		Id:       uint32(id),
		AgentKey: req.AgentKey,
	}, nil
}

func (s *Grpc) ListAgents(ctx context.Context, req *ListRequest) (*ListAgentsResponse, error) {
	page := util.NewPaginator(int(req.PageSize), int(req.PageNumber), req.SortBy)

	filter := util.NewFilter(req.SearchQuery)

	agents, total, err := agentService.ListAgents(page, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch agents: %v", err)
	}
	return convertToAgentResponse(agents, total)
}

func (s *Grpc) UpdateAgentGroup(ctx context.Context, req *AgentGroupUpdate) (*Agent, error) {
	if req.AgentId == 0 || req.AgentGroup == 0 {
		return nil, status.Errorf(codes.FailedPrecondition, "error in req")
	}
	agent, err := agentService.UpdateAgentGroup(uint(req.AgentId), uint(req.AgentGroup))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to update group: %v", err)
	}
	return parseAgentToProto(agent), nil
}

func (s *Grpc) GetAgentByHostname(ctx context.Context, req *Hostname) (*Agent, error) {
	if req.Hostname == "" {
		return nil, status.Errorf(codes.FailedPrecondition, "error in req")
	}
	agent, err := agentService.FindByHostname(req.Hostname)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "unable to find agent with hostname: %v", err)
	}
	return parseAgentToProto(*agent), nil
}

func (s *Grpc) UpdateAgentType(ctx context.Context, req *AgentTypeUpdate) (*Agent, error) {
	if req.AgentId == 0 || req.AgentType == 0 {
		return nil, status.Errorf(codes.FailedPrecondition, "error in req")
	}
	agent, err := agentService.UpdateAgentType(uint(req.AgentId), uint(req.AgentType))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to update type: %v", err)
	}
	return parseAgentToProto(agent), nil
}

func (s *Grpc) LoadAgentCacheFromDatabase() error {
	// Replace with your database loading logic
	// Fill the agentCache map with agentID and agentToken pairs
	agents, err := agentService.FindAll()
	if err != nil {
		log.Fatalf("Failed to fetch agents from database: %v", err)
		return err
	}
	for _, agent := range agents {
		Cache[agent.ID] = agent.AgentKey
	}
	return nil
}

func (s *Grpc) ListAgentsWithCommands(ctx context.Context, req *ListRequest) (*ListAgentsResponse, error) {
	page := util.NewPaginator(int(req.PageSize), int(req.PageNumber), req.SortBy)

	filter := util.NewFilter(req.SearchQuery)

	agents, total, err := agentService.ListAgentWithCommands(page, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch agents: %v", err)
	}

	// Return the list of agents as a ListAgentsCommandsResponse
	return convertToAgentResponse(agents, total)
}

func convertToAgentResponse(agents []models.Agent, total int64) (*ListAgentsResponse, error) {
	var agentMessages []*Agent
	for _, agent := range agents {
		agentProto := parseAgentToProto(agent)
		agentMessages = append(agentMessages, agentProto)
	}
	// Return the list of agents as a ListAgentsResponse
	return &ListAgentsResponse{
		Rows:  agentMessages,
		Total: int32(total),
	}, nil
}

// Wait for the stream to become ready again
func waitForStream(ctx context.Context, stream grpc.ServerStream) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context canceled: %v", ctx.Err())
		default:
			err := stream.Context().Err()
			if err != nil {
				if err == context.Canceled {
					return fmt.Errorf("context canceled: %v", err)
				}
				return fmt.Errorf("stream error: %v", err)
			}
			time.Sleep(time.Second)
		}
	}
}

// Wait for the client to reconnect
func waitForReconnect(ctx context.Context, s *Grpc) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context canceled: %v", ctx.Err())
		default:
			for _, stream := range s.AgentStreamMap {
				// Check if the stream is still valid
				err := stream.stream.Context().Err()
				if err == nil {
					// Stream is still valid, wait for a moment and try again
					time.Sleep(time.Second)
					break
				}
			}
			return nil
		}
	}
}

func createHistoryCommand(cmd *UtmCommand, cmdID string) {
	cmdHistory := &models.AgentCommand{
		AgentID:       findAgentIdByToken(Cache, cmd.AgentKey),
		Command:       cmd.Command,
		CommandStatus: models.Pending,
		Result:        "",
		ExecutedBy:    cmd.ExecutedBy,
		CmdId:         cmdID,
		OriginType:    cmd.OriginType,
		OriginId:      cmd.OriginId,
		Reason:        cmd.Reason,
	}
	err := agentCommandService.Create(cmdHistory)
	if err != nil {
		log.Printf("Unable to create a new command history")
	}
}

func updateHistoryCommand(cmdResult *CommandResult, cmdID string) {
	err := agentCommandService.UpdateCommandStatusAndResult(findAgentIdByToken(Cache, cmdResult.AgentKey), cmdID, models.Executed, cmdResult.Result)
	if err != nil {
		log.Printf("Failed to update command status")
	}
}

func parseAgentToProto(agent models.Agent) *Agent {
	agentStatus, lastSeen := agentService.GetAgentStatus(agent)
	return &Agent{
		Id:             uint32(agent.ID),
		Ip:             agent.Ip,
		Status:         AgentStatus(agentStatus),
		Hostname:       agent.Hostname,
		Os:             agent.Os,
		Platform:       agent.Platform,
		Version:        agent.Version,
		AgentKey:       agent.AgentKey,
		LastSeen:       lastSeen,
		Aliases:        agent.Aliases,
		Addresses:      agent.Addresses,
		Mac:            agent.Mac,
		OsMajorVersion: agent.OsMajorVersion,
		OsMinorVersion: agent.OsMinorVersion,
	}
}

func setAgentType(agentType *models.AgentType) *AgentType {
	if agentType != nil {
		return &AgentType{
			Id:       uint32(agentType.ID),
			TypeName: agentType.TypeName,
		}
	}
	return nil
}

func setAgentGroup(group *models.AgentGroup) *AgentGroup {
	if group != nil {
		return &AgentGroup{
			Id:               uint32(group.Id),
			GroupName:        group.GroupName,
			GroupDescription: group.GroupDescription,
		}
	}
	return nil
}
