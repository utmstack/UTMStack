package agent

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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
			util.Logger.ErrorF("Agent with hostname %s already exists", agent.Hostname)
			return nil, status.Errorf(codes.AlreadyExists, "hostname has already been registered")
		}
	}

	key := uuid.New().String()
	agent.AgentKey = key
	err = agentService.Create(agent)
	if err != nil {
		util.Logger.ErrorF("Failed to create agent: %v", err)
		return nil, err
	}

	s.cacheAgentMutex.Lock()
	CacheAgent[agent.ID] = key
	s.cacheAgentMutex.Unlock()

	err = lastSeenService.Set(key, time.Now())
	if err != nil {
		util.Logger.ErrorF("Failed to set last seen: %v", err)
		return nil, err
	}
	res := &AuthResponse{
		Id:  uint32(agent.ID),
		Key: key,
	}

	util.Logger.Info("Agent %s with id %d registered correctly", agent.Hostname, agent.ID)
	return res, nil
}

func (s *Grpc) UpdateAgent(ctx context.Context, req *AgentRequest) (*AuthResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return &AuthResponse{}, status.Error(codes.Internal, "unable to get metadata from context")
	}

	ids, ok := md["id"]
	if !ok || len(ids) == 0 {
		return &AuthResponse{}, status.Error(codes.Internal, "unable to get agent id from metadata")
	}

	id, err := strconv.Atoi(ids[0])
	if err != nil {
		return &AuthResponse{}, status.Error(codes.Internal, "unable to parse agent id from metadata")
	}

	agent, err := agentService.FindByID(uint(id))
	if err != nil {
		util.Logger.ErrorF("Failed to find agent: %v", err)
		return nil, err
	}

	if req.GetIp() != "" {
		agent.Ip = req.GetIp()
	}
	if req.GetHostname() != "" {
		agent.Hostname = req.GetHostname()
	}
	if req.GetVersion() != "" {
		agent.Version = req.GetVersion()
	}
	if req.GetMac() != "" {
		agent.Mac = req.GetMac()
	}
	if req.GetOsMajorVersion() != "" {
		agent.OsMajorVersion = req.GetOsMajorVersion()
	}
	if req.GetOsMinorVersion() != "" {
		agent.OsMinorVersion = req.GetOsMinorVersion()
	}
	if req.GetAliases() != "" {
		agent.Aliases = req.GetAliases()
	}
	if req.GetAddresses() != "" {
		agent.Addresses = req.GetAddresses()
	}

	err = agentService.Update(agent)
	if err != nil {
		util.Logger.ErrorF("Failed to update agent: %v", err)
		return nil, err
	}

	res := &AuthResponse{
		Id:  uint32(agent.ID),
		Key: agent.AgentKey,
	}

	util.Logger.Info("Agent %s with id %d updated correctly", agent.Hostname, agent.ID)
	return res, nil
}

func (s *Grpc) DeleteAgent(ctx context.Context, req *AgentDelete) (*AuthResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return &AuthResponse{}, status.Error(codes.Internal, "unable to get metadata from context")
	}

	keys, ok := md["key"]
	if !ok || len(keys) == 0 {
		return &AuthResponse{}, status.Error(codes.Internal, "unable to get agent key from metadata")
	}
	key := keys[0]

	id, err := agentService.Delete(uuid.MustParse(key), req.DeletedBy)
	if err != nil {
		util.Logger.ErrorF("Unable to delete agent: %v", err)
		return &AuthResponse{}, status.Error(codes.Internal, fmt.Sprintf("unable to delete agent: %v", err.Error()))
	}

	s.cacheAgentMutex.Lock()
	delete(CacheAgent, id)
	s.cacheAgentMutex.Unlock()

	s.agentStreamMutex.Lock()
	delete(s.AgentStreamMap, key)
	s.agentStreamMutex.Unlock()

	util.Logger.Info("Agent with key %s deleted by %s", key, req.DeletedBy)

	return &AuthResponse{
		Id:  uint32(id),
		Key: key,
	}, nil
}

func (s *Grpc) ListAgents(ctx context.Context, req *ListRequest) (*ListAgentsResponse, error) {
	page := util.NewPaginator(int(req.PageSize), int(req.PageNumber), req.SortBy)

	filter := util.NewFilter(req.SearchQuery)

	agents, total, err := agentService.ListAgents(page, filter)
	if err != nil {
		util.Logger.ErrorF("failed to fetch agents: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch agents: %v", err)
	}
	return convertToAgentResponse(agents, total)
}

func (s *Grpc) AgentStream(stream AgentService_AgentStreamServer) error {
	agentKey, err := s.authenticateConnector(stream, ConnectorType_AGENT)
	if err != nil {
		return err
	}

	s.agentStreamMutex.Lock()
	if _, ok := s.AgentStreamMap[agentKey]; ok {
		s.agentStreamMutex.Unlock()
		return fmt.Errorf("agent already connected")
	}
	s.AgentStreamMap[agentKey] = stream
	s.agentStreamMutex.Unlock()

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			err = s.waitForReconnect(s.AgentStreamMap[agentKey].Context(), agentKey, ConnectorType_AGENT)
			if err != nil {
				s.agentStreamMutex.Lock()
				delete(s.AgentStreamMap, agentKey)
				s.agentStreamMutex.Unlock()

				util.Logger.ErrorF("failed to reconnect to client: %v", err)
				return fmt.Errorf("failed to reconnect to client: %v", err)
			}

			agentKey, err = s.authenticateConnector(stream, ConnectorType_AGENT)
			if err != nil {
				s.agentStreamMutex.Lock()
				delete(s.AgentStreamMap, agentKey)
				s.agentStreamMutex.Unlock()

				return err
			}

			continue
		}
		if err != nil {
			s.agentStreamMutex.Lock()
			delete(s.AgentStreamMap, agentKey)
			s.agentStreamMutex.Unlock()
			return err
		}

		switch msg := in.StreamMessage.(type) {
		case *BidirectionalStream_Command:
			util.Logger.Info("Received command: %s", msg.Command.CmdId)

		case *BidirectionalStream_Result:
			util.Logger.Info("Received command result: %s", msg.Result.CmdId)

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
				util.Logger.ErrorF("Failed to send result to server: %v", err)
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
				util.Logger.ErrorF("Failed to find result channel for CmdID: %s", cmdID)
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
		decryptedValue, _ := util.DecryptValue(encryptedValue)
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
		util.Logger.ErrorF("Error in req")
		return nil, status.Errorf(codes.FailedPrecondition, "error in req")
	}
	agent, err := agentService.UpdateAgentGroup(uint(req.AgentId), uint(req.AgentGroup))
	if err != nil {
		util.Logger.ErrorF("Unable to update group: %v", err)
		return nil, status.Errorf(codes.Internal, "unable to update group: %v", err)
	}
	return parseAgentToProto(agent), nil
}

func (s *Grpc) GetAgentByHostname(ctx context.Context, req *Hostname) (*Agent, error) {
	if req.Hostname == "" {
		util.Logger.ErrorF("Error in req")
		return nil, status.Errorf(codes.FailedPrecondition, "error in req")
	}
	agent, err := agentService.FindByHostname(req.Hostname)
	if err != nil {
		util.Logger.ErrorF("Unable to find agent with hostname: %v", err)
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
		util.Logger.ErrorF("Unable to update type: %v", err)
		return nil, status.Errorf(codes.Internal, "unable to update type: %v", err)
	}
	return parseAgentToProto(agent), nil
}

func (s *Grpc) LoadAgentCacheFromDatabase() error {
	// Replace with your database loading logic
	// Fill the agentCache map with agentID and agentToken pairs
	agents, err := agentService.FindAll()
	if err != nil {
		util.Logger.ErrorF("Failed to fetch agents from database: %v", err)
		return err
	}
	for _, agent := range agents {
		CacheAgent[agent.ID] = agent.AgentKey
	}
	return nil
}

func (s *Grpc) ListAgentsWithCommands(ctx context.Context, req *ListRequest) (*ListAgentsResponse, error) {
	page := util.NewPaginator(int(req.PageSize), int(req.PageNumber), req.SortBy)

	filter := util.NewFilter(req.SearchQuery)

	agents, total, err := agentService.ListAgentWithCommands(page, filter)
	if err != nil {
		util.Logger.ErrorF("failed to fetch agents: %v", err)
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

/*
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
*/

func createHistoryCommand(cmd *UtmCommand, cmdID string) {
	cmdHistory := &models.AgentCommand{
		AgentID:       findAgentIdByKey(CacheAgent, cmd.AgentKey),
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
		util.Logger.ErrorF("Unable to create a new command history")
	}
}

func updateHistoryCommand(cmdResult *CommandResult, cmdID string) {
	err := agentCommandService.UpdateCommandStatusAndResult(findAgentIdByKey(CacheAgent, cmdResult.AgentKey), cmdID, models.Executed, cmdResult.Result)
	if err != nil {
		util.Logger.ErrorF("Failed to update command status")
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

/*
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
*/
