package agent

import (
	"context"
	"fmt"
	"io"
	"strconv"
	sync "sync"
	"time"

	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/agent-manager/database"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	AgentServ     *AgentService
	agentServOnce sync.Once
)

type AgentService struct {
	UnimplementedAgentServiceServer
	UnimplementedPanelServiceServer

	AgentStreamMap        map[uint]AgentService_AgentStreamServer
	AgentStreamMutex      sync.Mutex
	CacheAgentKey         map[uint]string
	CacheAgentKeyMutex    sync.Mutex
	CommandResultChannel  map[string]chan *CommandResult
	CommandResultChannelM sync.Mutex

	DBConnection *database.DB
}

func InitAgentService() error {
	var err error
	agentServOnce.Do(func() {
		AgentServ = &AgentService{
			AgentStreamMap:       make(map[uint]AgentService_AgentStreamServer),
			CacheAgentKey:        make(map[uint]string),
			CommandResultChannel: make(map[string]chan *CommandResult),
			DBConnection:         database.GetDB(),
		}

		agents := []models.Agent{}
		_, err = AgentServ.DBConnection.GetAll(&agents, "")
		if err != nil {
			err = fmt.Errorf("failed to fetch agents: %v", err)
			return
		}
		for _, agent := range agents {
			AgentServ.CacheAgentKey[agent.ID] = agent.AgentKey
		}
	})
	return err
}

func (s *AgentService) RegisterAgent(ctx context.Context, req *AgentRequest) (*AuthResponse, error) {
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

	oldAgent := &models.Agent{}
	err := s.DBConnection.GetFirst(oldAgent, "hostname = ?", agent.Hostname)
	if err == nil {
		if oldAgent.Ip == agent.Ip {
			return &AuthResponse{
				Id:  uint32(oldAgent.ID),
				Key: oldAgent.AgentKey,
			}, nil
		} else {
			utils.ALogger.ErrorF("agent with hostname %s already exists", agent.Hostname)
			return nil, status.Errorf(codes.AlreadyExists, "hostname has already been registered")
		}
	}

	key := uuid.New().String()
	agent.AgentKey = key
	err = s.DBConnection.Create(agent)
	if err != nil {
		utils.ALogger.ErrorF("failed to create agent: %v", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create agent: %v", err))
	}

	s.CacheAgentKeyMutex.Lock()
	s.CacheAgentKey[agent.ID] = key
	s.CacheAgentKeyMutex.Unlock()

	LastSeenChannel <- models.LastSeen{
		ConnectorType: "agent",
		ConnectorID:   agent.ID,
		LastPing:      time.Now(),
	}

	utils.ALogger.Info("Agent %s with id %d registered correctly", agent.Hostname, agent.ID)
	return &AuthResponse{
		Id:  uint32(agent.ID),
		Key: key,
	}, nil
}

func (s *AgentService) DeleteAgent(ctx context.Context, req *DeleteRequest) (*AuthResponse, error) {
	id, key, _, err := utils.GetItemsFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid context")
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	err = s.DBConnection.Upsert(&models.Agent{}, "id = ?", map[string]interface{}{"deleted_by": req.DeletedBy}, id)
	if err != nil {
		utils.ALogger.ErrorF("unable to update delete_by field in agent: %v", err)
	}

	err = s.DBConnection.Delete(&models.AgentCommand{}, "agent_id = ?", false, uint(idInt))
	if err != nil {
		utils.ALogger.ErrorF("unable to delete agent commands: %v", err)
		return &AuthResponse{}, status.Error(codes.Internal, fmt.Sprintf("unable to delete agent commands: %v", err.Error()))
	}

	err = s.DBConnection.Delete(&models.Agent{}, "id = ?", false, id)
	if err != nil {
		utils.ALogger.ErrorF("unable to delete agent: %v", err)
		return &AuthResponse{}, status.Error(codes.Internal, fmt.Sprintf("unable to delete agent: %v", err.Error()))
	}

	s.CacheAgentKeyMutex.Lock()
	delete(s.CacheAgentKey, uint(idInt))
	s.CacheAgentKeyMutex.Unlock()

	s.AgentStreamMutex.Lock()
	delete(s.AgentStreamMap, uint(idInt))
	s.AgentStreamMutex.Unlock()

	utils.ALogger.Info("Agent with key %s deleted by %s", key, req.DeletedBy)

	return &AuthResponse{
		Id:  uint32(idInt),
		Key: key,
	}, nil
}

func (s *AgentService) ListAgents(ctx context.Context, req *ListRequest) (*ListAgentsResponse, error) {
	page := utils.NewPaginator(int(req.PageSize), int(req.PageNumber), req.SortBy)
	filter := utils.NewFilter(req.SearchQuery)

	agents := []models.Agent{}
	total, err := s.DBConnection.GetByPagination(&agents, page, filter, "", false)
	if err != nil {
		utils.ALogger.ErrorF("failed to fetch agents: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch agents: %v", err)
	}

	return convertModelToAgentResponse(agents, total), nil
}

func (s *AgentService) AgentStream(stream AgentService_AgentStreamServer) error {
	id, _, _, err := utils.GetItemsFromContext(stream.Context())
	if err != nil {
		return err
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid id")
	}
	idUint := uint(idInt)

	s.AgentStreamMutex.Lock()
	if _, ok := s.AgentStreamMap[idUint]; ok {
		s.AgentStreamMutex.Unlock()
		return status.Error(codes.AlreadyExists, "stream already exists")
	}
	s.AgentStreamMap[idUint] = stream
	s.AgentStreamMutex.Unlock()

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			err = utils.WaitForReconnect(stream.Context(), stream)
			if err != nil {
				s.AgentStreamMutex.Lock()
				delete(s.AgentStreamMap, idUint)
				s.AgentStreamMutex.Unlock()

				return status.Error(codes.Internal, fmt.Sprintf("failed to reconnect: %v", err))
			}
			continue
		}
		if err != nil {
			s.AgentStreamMutex.Lock()
			delete(s.AgentStreamMap, idUint)
			s.AgentStreamMutex.Unlock()
			return status.Error(codes.Internal, fmt.Sprintf("failed to receive message: %v", err))
		}

		switch msg := in.StreamMessage.(type) {
		case *BidirectionalStream_Result:
			utils.ALogger.Info("Received command result from agent %s: %s", msg.Result.AgentId, msg.Result.Result)
			cmdID := msg.Result.GetCmdId()

			s.CommandResultChannelM.Lock()
			if resultChan, ok := s.CommandResultChannel[cmdID]; ok {
				resultChan <- &CommandResult{
					AgentId:    msg.Result.AgentId,
					Result:     msg.Result.Result,
					CmdId:      cmdID,
					ExecutedAt: msg.Result.ExecutedAt,
				}
			} else {
				utils.ALogger.ErrorF("failed to find result channel for CmdID: %s", cmdID)
			}
			s.CommandResultChannelM.Unlock()
		}
	}
}

func (s *AgentService) ProcessCommand(stream PanelService_ProcessCommandServer) error {
	for {
		cmd, err := stream.Recv()
		if err == io.EOF {
			return status.Error(codes.Internal, "stream closed")
		}
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("failed to receive message: %v", err))
		}
		streamId, err := strconv.Atoi(cmd.AgentId)
		if err != nil {
			return status.Error(codes.InvalidArgument, "invalid agent ID")
		}
		agentStream, ok := s.AgentStreamMap[uint(streamId)]
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

		cmdID := cmd.GetCmdId()
		if cmdID == "" {
			cmdID = uuid.New().String()
		}

		s.CommandResultChannelM.Lock()
		s.CommandResultChannel[cmdID] = make(chan *CommandResult)
		s.CommandResultChannelM.Unlock()

		histCommand := createHistoryCommand(cmd, cmdID, uint(streamId))
		err = s.DBConnection.Create(&histCommand)
		if err != nil {
			utils.ALogger.ErrorF("unable to create a new command history")
		}

		err = agentStream.Send(&BidirectionalStream{
			StreamMessage: &BidirectionalStream_Command{
				Command: &UtmCommand{
					AgentId: cmd.AgentId,
					Command: replaceSecretValues(cmd.Command),
					CmdId:   cmdID,
				},
			},
		})
		if err != nil {
			return status.Errorf(codes.Internal, "failed to send command to agent: %v", err)
		}

		result := <-s.CommandResultChannel[cmdID]
		err = s.DBConnection.Upsert(
			&models.AgentCommand{},
			"agent_id = ? AND cmd_id = ?",
			map[string]interface{}{"command_status": models.Executed, "result": result.Result},
			cmd.AgentId, cmdID,
		)
		if err != nil {
			utils.ALogger.ErrorF("failed to update command status: %v", err)
		}

		err = stream.Send(result)
		if err != nil {
			return err
		}

		s.CommandResultChannelM.Lock()
		delete(s.CommandResultChannel, cmdID)
		s.CommandResultChannelM.Unlock()
	}
}

func (s *AgentService) ListAgentCommands(ctx context.Context, req *ListRequest) (*ListAgentsCommandsResponse, error) {
	page := utils.NewPaginator(int(req.PageSize), int(req.PageNumber), req.SortBy)
	filter := utils.NewFilter(req.SearchQuery)

	commands := []models.AgentCommand{}
	total, err := s.DBConnection.GetByPagination(&commands, page, filter, "", false)
	if err != nil {
		utils.ALogger.ErrorF("failed to fetch agent commands: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch agent commands: %v", err)
	}

	return &ListAgentsCommandsResponse{
		Rows:  convertModelToAgentCommandsProto(commands),
		Total: int32(total),
	}, nil
}
