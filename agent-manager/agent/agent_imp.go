package agent

import (
	"context"
	"fmt"
	"io"
	"strconv"
	sync "sync"
	"time"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/agent-manager/authcache"
	"github.com/utmstack/UTMStack/agent-manager/database"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	AgentServ     *AgentService
	agentServOnce sync.Once

	// AuthCache publishes keys where the log input reads them. Nil until
	// configured, and every call on a nil publisher is a no-op.
	AuthCache *authcache.Publisher
)

type AgentService struct {
	UnimplementedAgentServiceServer
	UnimplementedPanelServiceServer

	AgentStreamMap        map[uint]AgentService_AgentStreamServer
	AgentStreamMutex      sync.Mutex
	CacheAgentKey         map[uint]utils.ConnectorAuth
	CacheAgentKeyMutex    sync.RWMutex
	CommandResultChannel  map[string]chan *CommandResult
	CommandResultChannelM sync.Mutex
	connKeys              map[string]models.ConnectionKey // tenant -> key
	connKeyMutex          sync.RWMutex
	DBConnection          *database.DB
}

func (s *AgentService) ValidateAgentKey(key string, id uint) bool {
	s.CacheAgentKeyMutex.RLock()
	defer s.CacheAgentKeyMutex.RUnlock()
	_, valid := utils.IsKeyPairValid(key, id, s.CacheAgentKey)
	return valid
}

func InitAgentService() error {
	var err error
	agentServOnce.Do(func() {
		AgentServ = &AgentService{
			AgentStreamMap:       make(map[uint]AgentService_AgentStreamServer),
			CacheAgentKey:        make(map[uint]utils.ConnectorAuth),
			CommandResultChannel: make(map[string]chan *CommandResult),
			connKeys:             make(map[string]models.ConnectionKey),
			DBConnection:         database.GetDB(),
		}

		agents := []models.Agent{}
		_, err = AgentServ.DBConnection.GetAll(&agents, "")
		if err != nil {
			err = fmt.Errorf("failed to fetch agents: %v", err)
			return
		}
		for _, agent := range agents {
			AgentServ.CacheAgentKey[agent.ID] = utils.ConnectorAuth{Key: agent.AgentKey, TenantID: tenantOrDefault(agent.TenantID)}
		}

		if e := AgentServ.loadConnectionKeys(); e != nil {
			err = e
			return
		}
	})
	return err
}

func (s *AgentService) RegisterAgent(ctx context.Context, req *AgentRequest) (*AuthResponse, error) {
	tenantID, ok := s.tenantFromEnrolment(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "connection key does not belong to a tenant")
	}

	agent := &models.Agent{
		TenantID:       tenantID,
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
	err := s.DBConnection.GetFirst(oldAgent, "hostname = ? AND mac = ?", agent.Hostname, agent.Mac)
	if err == nil {
		// Same machine re-registering, return existing agent
		return &AuthResponse{
			Id:  uint32(oldAgent.ID),
			Key: oldAgent.AgentKey,
		}, nil
	}

	key := uuid.New().String()
	agent.AgentKey = key
	err = s.DBConnection.Create(agent)
	if err != nil {
		catcher.Error("failed to create agent", err, map[string]any{"process": "agent-manager"})
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create agent: %v", err))
	}

	s.CacheAgentKeyMutex.Lock()
	entry := utils.ConnectorAuth{Key: key, TenantID: tenantOrDefault(agent.TenantID)}
	s.CacheAgentKey[agent.ID] = entry
	AuthCache.PublishAgent(agent.ID, entry)
	s.CacheAgentKeyMutex.Unlock()

	LastSeenChannel <- models.LastSeen{
		ConnectorType: "agent",
		ConnectorID:   agent.ID,
		LastPing:      time.Now(),
	}

	catcher.Info("Agent registered correctly", map[string]any{"hostname": agent.Hostname, "id": agent.ID, "process": "agent-manager"})

	if OnAgentRegisterHook != nil {
		OnAgentRegisterHook(agent)
	}

	return &AuthResponse{
		Id:  uint32(agent.ID),
		Key: key,
	}, nil
}

func (s *AgentService) UpdateAgent(ctx context.Context, req *AgentRequest) (*AuthResponse, error) {
	id, key, _, err := utils.GetItemsFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid context")
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	agent := &models.Agent{}
	err = s.DBConnection.GetFirst(agent, "id = ?", idInt)
	if err != nil {
		catcher.Error("failed to fetch agent", err, map[string]any{"process": "agent-manager"})
		return nil, status.Errorf(codes.NotFound, "agent not found")
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

	err = s.DBConnection.Upsert(&agent, "id = ?", nil, idInt)
	if err != nil {
		catcher.Error("failed to update agent", err, map[string]any{"process": "agent-manager"})
		return nil, status.Errorf(codes.Internal, "failed to update agent: %v", err)
	}

	if OnAgentUpdateHook != nil {
		OnAgentUpdateHook(agent)
	}

	res := &AuthResponse{
		Id:  uint32(agent.ID),
		Key: key,
	}

	return res, nil
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
		catcher.Error("unable to update delete_by field in agent", err, map[string]any{"process": "agent-manager"})
	}

	err = s.DBConnection.Delete(&models.AgentCommand{}, "agent_id = ?", false, uint(idInt))
	if err != nil {
		catcher.Error("unable to delete agent commands", err, map[string]any{"process": "agent-manager"})
		return &AuthResponse{}, status.Error(codes.Internal, fmt.Sprintf("unable to delete agent commands: %v", err.Error()))
	}

	err = s.DBConnection.Delete(&models.Agent{}, "id = ?", false, id)
	if err != nil {
		catcher.Error("unable to delete agent", err, map[string]any{"process": "agent-manager"})
		return &AuthResponse{}, status.Error(codes.Internal, fmt.Sprintf("unable to delete agent: %v", err.Error()))
	}

	s.CacheAgentKeyMutex.Lock()
	delete(s.CacheAgentKey, uint(idInt))
	AuthCache.DeleteAgent(uint(idInt))
	s.CacheAgentKeyMutex.Unlock()

	s.AgentStreamMutex.Lock()
	delete(s.AgentStreamMap, uint(idInt))
	s.AgentStreamMutex.Unlock()

	catcher.Info("Agent deleted", map[string]any{"key": key, "deleted_by": req.DeletedBy, "process": "agent-manager"})

	return &AuthResponse{
		Id:  uint32(idInt),
		Key: key,
	}, nil
}

func (s *AgentService) ListAgents(ctx context.Context, req *ListRequest) (*ListAgentsResponse, error) {
	page := utils.NewPaginator(int(req.PageSize), int(req.PageNumber), req.SortBy)
	filter := utils.NewFilter(req.SearchQuery)

	// Scoped by the caller: the panel asks for one tenant, and only the
	// platform asks for all of them.
	if req.GetTenantId() != "" {
		filter = append(filter, utils.Filter{
			Field: "tenant_id",
			Op: utils.Is,
			Value:sanitizeTenant(req.GetTenantId()),
		})
	}

	agents := []models.Agent{}
	total, err := s.DBConnection.GetByPagination(&agents, page, filter, "", false)
	if err != nil {
		catcher.Error("failed to fetch agents", err, map[string]any{"process": "agent-manager"})
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

	if OnAgentConnectHook != nil {
		OnAgentConnectHook(stream.Context(), idUint)
	}

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
			catcher.Info("Received command result from agent", map[string]any{"agent_id": msg.Result.AgentId, "result": msg.Result.Result, "process": "agent-manager"})
			cmdID := msg.Result.GetCmdId()

			s.CommandResultChannelM.Lock()
			if resultChan, ok := s.CommandResultChannel[cmdID]; ok {
				resultChan <- &CommandResult{
					AgentId:    msg.Result.AgentId,
					Result:     msg.Result.Result,
					CmdId:      cmdID,
					ExecutedAt: msg.Result.ExecutedAt,
				}
			} else if OnCommandResultHook == nil || !OnCommandResultHook(msg.Result) {
				catcher.Error("failed to find result channel for CmdID", nil, map[string]any{"cmdID": cmdID, "process": "agent-manager"})
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
			catcher.Error("unable to create a new command history", err, map[string]any{"process": "agent-manager"})
		}

		var lock sync.Locker
		if LockStreamHook != nil {
			lock = LockStreamHook(uint(streamId))
		}
		func() {
			if lock != nil {
				lock.Lock()
				defer lock.Unlock()
			}
			err = agentStream.Send(&BidirectionalStream{
				StreamMessage: &BidirectionalStream_Command{
					Command: &UtmCommand{
						AgentId: cmd.AgentId,
						Command: replaceSecretValues(cmd.Command),
						CmdId:   cmdID,
						Shell:   cmd.Shell,
					},
				},
			})
		}()
		if err != nil {
			return status.Errorf(codes.Internal, "failed to send command to agent: %v", err)
		}

		select {
		case result := <-s.CommandResultChannel[cmdID]:
			err = s.DBConnection.Upsert(
				&models.AgentCommand{},
				"agent_id = ? AND cmd_id = ?",
				map[string]interface{}{"command_status": models.Executed, "result": result.Result},
				cmd.AgentId, cmdID,
			)
			if err != nil {
				catcher.Error("failed to update command status", err, map[string]any{"process": "agent-manager"})
			}

			err = stream.Send(result)
			if err != nil {
				return err
			}
		case <-time.After(5 * time.Minute):
			s.CommandResultChannelM.Lock()
			delete(s.CommandResultChannel, cmdID)
			s.CommandResultChannelM.Unlock()

			_ = s.DBConnection.Upsert(
				&models.AgentCommand{},
				"agent_id = ? AND cmd_id = ?",
				map[string]interface{}{"command_status": models.Error, "result": "command timed out after 5 minutes"},
				cmd.AgentId, cmdID,
			)

			return status.Errorf(codes.DeadlineExceeded, "agent did not respond within 5 minutes")
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
		catcher.Error("failed to fetch agent commands", err, map[string]any{"process": "agent-manager"})
		return nil, status.Errorf(codes.Internal, "failed to fetch agent commands: %v", err)
	}

	return &ListAgentsCommandsResponse{
		Rows:  convertModelToAgentCommandsProto(commands),
		Total: int32(total),
	}, nil
}

// tenantFromEnrolment resolves which tenant the presented connection key
// enrols into. The interceptor has already accepted the key; this is what says
// whose it is.
func (s *AgentService) tenantFromEnrolment(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}
	keys := md.Get("connection-key")
	if len(keys) == 0 {
		return "", false
	}
	return s.TenantForConnectionKey(keys[0])
}

// sanitizeTenant keeps a tenant id to the shape an id has, because it is
// interpolated into a where clause rather than bound.
func sanitizeTenant(id string) string {
	out := make([]rune, 0, len(id))
	for _, r := range id {
		if r == '-' || (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			out = append(out, r)
		}
	}
	return string(out)
}

// AuthSnapshot is every key that should be published, used to heal an evicted
// or restarted cache.
func AuthSnapshot() authcache.Snapshot {
	s := authcache.Snapshot{
		Agents:     make(map[uint]utils.ConnectorAuth),
		Collectors: make(map[uint]utils.ConnectorAuth),
	}

	if AgentServ != nil {
		AgentServ.CacheAgentKeyMutex.RLock()
		for id, key := range AgentServ.CacheAgentKey {
			s.Agents[id] = key
		}
		AgentServ.CacheAgentKeyMutex.RUnlock()
	}

	if CollectorServ != nil {
		CollectorServ.CacheCollectorKeyMutex.RLock()
		for id, key := range CollectorServ.CacheCollectorKey {
			s.Collectors[id] = key
		}
		CollectorServ.CacheCollectorKeyMutex.RUnlock()
	}

	return s
}
