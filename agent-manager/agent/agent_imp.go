package agent

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Grpc) RegisterAgent(ctx context.Context, req *AgentRequest) (*AuthResponse, error) {
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
			return &AuthResponse{
				Id:  oldAgent.Id,
				Key: oldAgent.AgentKey,
			}, nil
		} else {
			utils.ALogger.ErrorF("agent with hostname %s already exists", agent.Hostname)
			return nil, status.Errorf(codes.AlreadyExists, "hostname has already been registered")
		}
	}

	key := uuid.New().String()
	agent.AgentKey = key
	err = agentService.Create(agent)
	if err != nil {
		utils.ALogger.ErrorF("failed to create agent: %v", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create agent: %v", err))
	}

	s.cacheAgentKeyMutex.Lock()
	CacheAgentKey[agent.ID] = key
	s.cacheAgentKeyMutex.Unlock()

	err = lastSeenService.Set(key, time.Now())
	if err != nil {
		utils.ALogger.ErrorF("failed to set last seen: %v", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to set last seen: %v", err))
	}
	res := &AuthResponse{
		Id:  uint32(agent.ID),
		Key: key,
	}

	utils.ALogger.Info("Agent %s with id %d registered correctly", agent.Hostname, agent.ID)
	return res, nil
}

func (s *Grpc) DeleteAgent(ctx context.Context, req *AgentDelete) (*AuthResponse, error) {
	key, err := getKeyFromContext(ctx)
	if err != nil {
		return nil, err
	}

	id, err := agentService.Delete(uuid.MustParse(key), req.DeletedBy)
	if err != nil {
		utils.ALogger.ErrorF("unable to delete agent: %v", err)
		return &AuthResponse{}, status.Error(codes.Internal, fmt.Sprintf("unable to delete agent: %v", err.Error()))
	}

	s.cacheAgentKeyMutex.Lock()
	delete(CacheAgentKey, id)
	s.cacheAgentKeyMutex.Unlock()

	s.agentStreamMutex.Lock()
	delete(s.AgentStreamMap, key)
	s.agentStreamMutex.Unlock()

	utils.ALogger.Info("Agent with key %s deleted by %s", key, req.DeletedBy)

	return &AuthResponse{
		Id:  uint32(id),
		Key: key,
	}, nil
}

func (s *Grpc) ListAgents(ctx context.Context, req *ListRequest) (*ListAgentsResponse, error) {
	page := utils.NewPaginator(int(req.PageSize), int(req.PageNumber), req.SortBy)

	filter := utils.NewFilter(req.SearchQuery)

	agents, total, err := agentService.ListAgents(page, filter)
	if err != nil {
		utils.ALogger.ErrorF("failed to fetch agents: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch agents: %v", err)
	}
	return convertToAgentResponse(agents, total), nil
}

func (s *Grpc) AgentStream(stream AgentService_AgentStreamServer) error {
	key, err := getKeyFromContext(stream.Context())
	if err != nil {
		return err
	}

	s.agentStreamMutex.Lock()
	if _, ok := s.AgentStreamMap[key]; ok {
		s.agentStreamMutex.Unlock()
		return fmt.Errorf("agent already connected")
	}
	s.AgentStreamMap[key] = stream
	s.agentStreamMutex.Unlock()

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			err = s.waitForReconnect(s.AgentStreamMap[key].Context(), key, ConnectorType_AGENT)
			if err != nil {
				s.agentStreamMutex.Lock()
				delete(s.AgentStreamMap, key)
				s.agentStreamMutex.Unlock()

				return fmt.Errorf("failed to reconnect to client: %v", err)
			}
			continue
		}
		if err != nil {
			s.agentStreamMutex.Lock()
			delete(s.AgentStreamMap, key)
			s.agentStreamMutex.Unlock()
			return err
		}

		switch msg := in.StreamMessage.(type) {
		case *BidirectionalStream_Command:
			utils.ALogger.Info("Received command: %s", msg.Command.CmdId)

		case *BidirectionalStream_Result:
			utils.ALogger.Info("Received command result: %s", msg.Result.CmdId)

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
				utils.ALogger.ErrorF("failed to send result to server: %v", err)
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
				utils.ALogger.ErrorF("failed to find result channel for CmdID: %s", cmdID)
			}
			s.resultChannelM.Unlock()
		}
	}
}

func (s *Grpc) replaceSecretValues(input string) string {
	pattern := regexp.MustCompile(`\$\[(\w+):([^]]+)]`)
	return pattern.ReplaceAllStringFunc(input, func(match string) string {
		matches := pattern.FindStringSubmatch(match)
		if len(matches) < 3 {
			return match // In case of no match, return the original
		}
		encryptedValue := matches[2]
		decryptedValue, _ := utils.DecryptValue(encryptedValue)
		return decryptedValue
	})
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
		err = agentStream.Send(&BidirectionalStream{
			StreamMessage: &BidirectionalStream_Command{
				Command: &UtmCommand{
					AgentKey: cmd.AgentKey,
					Command:  s.replaceSecretValues(cmd.Command),
					CmdId:    cmdID,
				},
			},
		})
		if err != nil {
			return err
		}

		// Wait for the result from the agent
		result := <-resultChan
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

func (s *Grpc) UpdateAgentGroup(ctx context.Context, req *AgentGroupUpdate) (*Agent, error) {
	if req.AgentId == 0 || req.AgentGroup == 0 {
		utils.ALogger.ErrorF("error in req")
		return nil, status.Errorf(codes.FailedPrecondition, "error in req")
	}
	agent, err := agentService.UpdateAgentGroup(uint(req.AgentId), uint(req.AgentGroup))
	if err != nil {
		utils.ALogger.ErrorF("unable to update group: %v", err)
		return nil, status.Errorf(codes.Internal, "unable to update group: %v", err)
	}
	return parseAgentToProto(agent), nil
}

func (s *Grpc) GetAgentByHostname(ctx context.Context, req *Hostname) (*Agent, error) {
	if req.Hostname == "" {
		utils.ALogger.ErrorF("error in req")
		return nil, status.Errorf(codes.FailedPrecondition, "error in req")
	}
	agent, err := agentService.FindByHostname(req.Hostname)
	if err != nil {
		utils.ALogger.ErrorF("unable to find agent with hostname: %s: %v", req.Hostname, err)
		return nil, status.Errorf(codes.NotFound, "unable to find agent with hostname: %s: %v", req.Hostname, err)
	}
	return parseAgentToProto(*agent), nil
}

func (s *Grpc) UpdateAgentType(ctx context.Context, req *AgentTypeUpdate) (*Agent, error) {
	if req.AgentId == 0 || req.AgentType == 0 {
		return nil, status.Errorf(codes.FailedPrecondition, "error in req")
	}
	agent, err := agentService.UpdateAgentType(uint(req.AgentId), uint(req.AgentType))
	if err != nil {
		utils.ALogger.ErrorF("unable to update type: %v", err)
		return nil, status.Errorf(codes.Internal, "unable to update type: %v", err)
	}
	return parseAgentToProto(agent), nil
}

func (s *Grpc) LoadAgentCacheFromDatabase() error {
	agents, err := agentService.FindAll()
	if err != nil {
		utils.ALogger.ErrorF("failed to fetch agents from database: %v", err)
		return err
	}
	for _, agent := range agents {
		CacheAgentKey[agent.ID] = agent.AgentKey
	}
	return nil
}

func (s *Grpc) ListAgentsWithCommands(ctx context.Context, req *ListRequest) (*ListAgentsResponse, error) {
	page := utils.NewPaginator(int(req.PageSize), int(req.PageNumber), req.SortBy)

	filter := utils.NewFilter(req.SearchQuery)

	agents, total, err := agentService.ListAgentWithCommands(page, filter)
	if err != nil {
		utils.ALogger.ErrorF("failed to fetch agents: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch agents: %v", err)
	}

	return convertToAgentResponse(agents, total), nil
}

func convertToAgentResponse(agents []models.Agent, total int64) *ListAgentsResponse {
	var agentMessages []*Agent
	for _, agent := range agents {
		agentProto := parseAgentToProto(agent)
		agentMessages = append(agentMessages, agentProto)
	}
	return &ListAgentsResponse{
		Rows:  agentMessages,
		Total: int32(total),
	}
}

func createHistoryCommand(cmd *UtmCommand, cmdID string) {
	cmdHistory := &models.AgentCommand{
		AgentID:       findAgentIdByKey(CacheAgentKey, cmd.AgentKey),
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
		utils.ALogger.ErrorF("unable to create a new command history")
	}
}

func updateHistoryCommand(cmdResult *CommandResult, cmdID string) {
	err := agentCommandService.UpdateCommandStatusAndResult(findAgentIdByKey(CacheAgentKey, cmdResult.AgentKey), cmdID, models.Executed, cmdResult.Result)
	if err != nil {
		utils.ALogger.ErrorF("failed to update command status")
	}
}

func parseAgentToProto(agent models.Agent) *Agent {
	agentStatus, lastSeen := lastSeenService.GetStatus(agent.AgentKey)
	return &Agent{
		Id:             uint32(agent.ID),
		Ip:             agent.Ip,
		Status:         Status(agentStatus),
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
