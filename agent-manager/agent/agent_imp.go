package agent

import (
	"context"
	"errors"
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
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

const agentHeartbeatInterval = 30 * time.Second

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

	configRegistry  map[string]map[uint]string // key -> agentID (0 = fleet-wide default) -> content
	configRevisions map[string]uint64          // key -> shared revision, see models.AgentConfigRevision
	configMutex     sync.RWMutex
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

		if e := AgentServ.loadConfigRegistry(); e != nil {
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
		TenantID:        tenantID,
		Ip:              req.GetIp(),
		Hostname:        req.GetHostname(),
		Os:              req.GetOs(),
		Platform:        req.GetPlatform(),
		Version:         req.GetVersion(),
		RegisterBy:      req.GetRegisterBy(),
		Mac:             req.GetMac(),
		OsMajorVersion:  req.GetOsMajorVersion(),
		OsMinorVersion:  req.GetOsMinorVersion(),
		Aliases:         req.GetAliases(),
		Addresses:       req.GetAddresses(),
		NoRemoteControl: req.GetNoRemoteControl(),
	}

	oldAgent := &models.Agent{}
	err := s.DBConnection.GetFirst(oldAgent, "hostname = ? AND mac = ?", agent.Hostname, agent.Mac)
	if err == nil {
		if oldAgent.NoRemoteControl != agent.NoRemoteControl {
			if _, uErr := s.DBConnection.UpdateOnly(&models.Agent{}, "id = ?",
				map[string]interface{}{"no_remote_control": agent.NoRemoteControl}, oldAgent.ID); uErr != nil {
				catcher.Error("failed to update the agent remote-control stance", uErr,
					map[string]any{"process": "agent-manager", "agent": oldAgent.ID})
			}
		}
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
			Op:    utils.Is,
			Value: sanitizeTenant(req.GetTenantId()),
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

func (s *AgentService) GetAgentAuth(ctx context.Context, req *ConnectorAuthRequest) (*ConnectorAuthResponse, error) {
	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	agent := models.Agent{}
	if err := s.DBConnection.GetFirst(&agent, "id = ?", req.GetId()); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "agent not found")
		}
		catcher.Error("failed to fetch agent", err, map[string]any{"process": "agent-manager"})
		return nil, status.Errorf(codes.Internal, "failed to fetch agent: %v", err)
	}

	return &ConnectorAuthResponse{Key: agent.AgentKey, TenantId: agent.TenantID}, nil
}

// evictIfOwner deletes the AgentStreamMap entry for agentID only if it still
// points to stream. Prevents a slow-exiting prior AgentStream goroutine from
// clobbering the fresh entry a newly-reconnected agent installed.
func (s *AgentService) evictIfOwner(agentID uint, stream AgentService_AgentStreamServer) {
	s.AgentStreamMutex.Lock()
	if s.AgentStreamMap[agentID] == stream {
		delete(s.AgentStreamMap, agentID)
	}
	s.AgentStreamMutex.Unlock()
}

// storeEdrStatus keeps the latest EDR status report per agent in the DB.
// Never deletes: an offline agent's last report stays readable.
func (s *AgentService) storeEdrStatus(agentID uint, report *EdrStatusReport) {
	tenant := s.agentTenant(agentID)
	reportedAt := time.Now().UTC()
	if ts := report.GetReportedAt(); ts != nil {
		reportedAt = ts.AsTime()
	}
	existing := &models.AgentEdrStatus{}
	err := s.DBConnection.GetFirst(existing, "agent_id = ?", agentID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			catcher.Error("failed to fetch EDR status report", err, map[string]any{"agent_id": agentID, "process": "agent-manager"})
			return
		}
		// First report for this agent: insert.
		rec := &models.AgentEdrStatus{
			AgentID:       agentID,
			TenantID:      tenant,
			StatusJSON:    report.GetStatusJson(),
			PolicyVersion: report.GetPolicyVersion(),
			ReportedAt:    reportedAt,
		}
		if err := s.DBConnection.Create(rec); err != nil {
			catcher.Error("failed to store EDR status report", err, map[string]any{"agent_id": agentID, "process": "agent-manager"})
		}
		return
	}
	if _, err := s.DBConnection.UpdateOnly(existing, "id = ?", map[string]interface{}{
		"tenant_id":      tenant,
		"status_json":    report.GetStatusJson(),
		"policy_version": report.GetPolicyVersion(),
		"reported_at":    reportedAt,
	}, existing.ID); err != nil {
		catcher.Error("failed to update EDR status report", err, map[string]any{"agent_id": agentID, "process": "agent-manager"})
	}
}

// agentTenant resolves the tenant of agentID for tenant-scoped reads: the
// enrolment key cache first, then the agents table. Returns "" when unknown.
func (s *AgentService) agentTenant(agentID uint) string {
	s.CacheAgentKeyMutex.Lock()
	tenant := s.CacheAgentKey[agentID].TenantID
	s.CacheAgentKeyMutex.Unlock()
	if tenant != "" {
		return tenant
	}
	agent := models.Agent{}
	if err := s.DBConnection.GetFirst(&agent, "id = ?", agentID); err == nil {
		return tenantOrDefault(agent.TenantID)
	}
	return ""
}

// loadEdrStatus returns the latest stored EDR status report for an agent.
// (nil, nil) when the agent never sent one, (nil, err) on a real DB error.
func (s *AgentService) loadEdrStatus(agentID uint) (*models.AgentEdrStatus, error) {
	rec := &models.AgentEdrStatus{}
	if err := s.DBConnection.GetFirst(rec, "agent_id = ?", agentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return rec, nil
}

// tenantGuard resolves the caller's tenant for scoping: the request payload
// wins (the internal-key path carries no tenant in context, same as ListAgents),
// then the auth context, else "" = platform view, no guard.
func tenantGuard(ctx context.Context, reqTenant string) string {
	if reqTenant != "" {
		return reqTenant
	}
	if tid, ok := tenantFromContext(ctx); ok {
		return tid
	}
	return ""
}

// GetEdrStatus returns stored EDR status reports. req.AgentId 0 means every
// agent the caller may see: tenant-scoped callers get their own rows, and a
// caller with no tenant (platform) gets all of them.
func (s *AgentService) GetEdrStatus(ctx context.Context, req *EdrStatusListRequest) (*EdrStatusListResponse, error) {
	if req.GetAgentId() != 0 {
		rec, err := s.loadEdrStatus(uint(req.GetAgentId()))
		if err != nil {
			catcher.Error("failed to fetch EDR status report", err, map[string]any{"agent_id": req.GetAgentId(), "process": "agent-manager"})
			return nil, status.Errorf(codes.Internal, "failed to fetch edr status: %v", err)
		}
		rows := &EdrStatusListResponse{}
		if rec != nil {
			rows.Rows = append(rows.Rows, edrStatusRow(rec))
		}
		return rows, nil
	}

	rows := []models.AgentEdrStatus{}
	var err error
	if tid := tenantGuard(ctx, req.GetTenantId()); tid != "" {
		_, err = s.DBConnection.GetAll(&rows, "tenant_id = ?", tid)
	} else {
		_, err = s.DBConnection.GetAll(&rows, "")
	}
	if err != nil {
		catcher.Error("failed to fetch EDR status reports", err, map[string]any{"process": "agent-manager"})
		return nil, status.Errorf(codes.Internal, "failed to fetch edr statuses: %v", err)
	}
	resp := &EdrStatusListResponse{}
	for i := range rows {
		resp.Rows = append(resp.Rows, edrStatusRow(&rows[i]))
	}
	return resp, nil
}

func edrStatusRow(rec *models.AgentEdrStatus) *EdrStatusRow {
	ts := timestamppb.New(rec.ReportedAt)
	if rec.ReportedAt.IsZero() {
		ts = nil
	}
	return &EdrStatusRow{
		AgentId:       uint32(rec.AgentID),
		StatusJson:    rec.StatusJSON,
		PolicyVersion: rec.PolicyVersion,
		ReportedAt:    ts,
	}
}

// storeEdrResult folds a returned EdrResult into the pending history row that
// dispatchOne created for the same cmd id. The proto carries no executed_by,
// so the row keeps the dispatcher's value; only status + result change.
func (s *AgentService) storeEdrResult(res *EdrResult) {
	agentID, err := strconv.ParseUint(res.GetAgentId(), 10, 64)
	if err != nil {
		catcher.Error("failed to parse edr result agent id", err, map[string]any{"agent_id": res.GetAgentId(), "process": "agent-manager"})
		return
	}
	statusStr := models.Executed
	if !res.GetOk() {
		statusStr = models.Error
	}
	result := res.GetPayload()
	if !res.GetOk() && res.GetError() != "" {
		if result == "" {
			result = res.GetError()
		} else {
			result = result + " (" + res.GetError() + ")"
		}
	}
	if uErr := s.DBConnection.Upsert(
		&models.AgentCommand{},
		"agent_id = ? AND cmd_id = ?",
		map[string]interface{}{"command_status": statusStr, "result": result},
		agentID, res.GetCmdId(),
	); uErr != nil {
		catcher.Error("failed to update edr command result", uErr, map[string]any{"cmd_id": res.GetCmdId(), "process": "agent-manager"})
	}
}

// DispatchEdrCommand pushes one EdrCommand to a connected agent and records it
// in the history table as pending. The result arrives asynchronously as an
// EdrResult on the AgentStream and is folded in by storeEdrResult.
func (s *AgentService) DispatchEdrCommand(ctx context.Context, req *DispatchEdrCommandRequest) (*DispatchEdrCommandResponse, error) {
	return s.dispatchOne(ctx, req.GetAgentId(), req.GetCmdId(), req.GetAction(), req.GetPayload(), req.GetExecutedBy(), req.GetReason(), tenantGuard(ctx, req.GetTenantId()))
}

// DispatchEdrCommands pushes the same action/payload to a set of agents, e.g.
// assigning a policy or an allowlist to a group. Per-target results are
// returned; a target that is offline is reported, not fatal to the batch.
func (s *AgentService) DispatchEdrCommands(ctx context.Context, req *DispatchEdrCommandsRequest) (*DispatchEdrCommandsResponse, error) {
	resp := &DispatchEdrCommandsResponse{}
	tenant := tenantGuard(ctx, req.GetTenantId())
	for _, t := range req.GetTargets() {
		one, err := s.dispatchOne(ctx, t.GetAgentId(), t.GetCmdId(), t.GetAction(), t.GetPayload(), req.GetExecutedBy(), req.GetReason(), tenant)
		if err != nil {
			resp.Rows = append(resp.Rows, &EdrCommandResultRow{
				AgentId:   t.GetAgentId(),
				CmdId:     t.GetCmdId(),
				Dispatched: false,
				Error:     err.Error(),
			})
			continue
		}
		resp.Rows = append(resp.Rows, &EdrCommandResultRow{
			AgentId:   t.GetAgentId(),
			CmdId:     one.GetCmdId(),
			Dispatched: one.GetDispatched(),
			Error:     "",
		})
	}
	return resp, nil
}

func (s *AgentService) dispatchOne(ctx context.Context, agentIDStr, cmdID, action, payload, executedBy, reason, tenant string) (*DispatchEdrCommandResponse, error) {
	agentID, err := strconv.ParseUint(agentIDStr, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid agent id: %s", agentIDStr)
	}
	if action == "" {
		return nil, status.Errorf(codes.InvalidArgument, "action is required")
	}
	if cmdID == "" {
		cmdID = uuid.New().String()
	}

	// Tenant guard: a tenant-scoped caller may only reach its own agents.
	if tenant != "" {
		if s.agentTenant(uint(agentID)) != tenant {
			return nil, status.Errorf(codes.PermissionDenied, "agent %s is not in caller tenant", agentIDStr)
		}
	}

	s.AgentStreamMutex.Lock()
	agentStream, connected := s.AgentStreamMap[uint(agentID)]
	s.AgentStreamMutex.Unlock()
	if !connected {
		return nil, status.Errorf(codes.NotFound, "agent %s is not connected", agentIDStr)
	}

	hist := &models.AgentCommand{
		AgentID:       uint(agentID),
		Command:       "edr: " + action,
		CommandStatus: models.Pending,
		Result:        "",
		ExecutedBy:    executedBy,
		CmdId:         cmdID,
		OriginType:    "edr",
		OriginId:      cmdID,
		Reason:        reason,
	}
	if cErr := s.DBConnection.Create(hist); cErr != nil {
		catcher.Error("unable to create edr command history", cErr, map[string]any{"process": "agent-manager"})
	}

	var lock sync.Locker
	if LockStreamHook != nil {
		lock = LockStreamHook(uint(agentID))
	}
	func() {
		if lock != nil {
			lock.Lock()
			defer lock.Unlock()
		}
		_ = agentStream.Send(&BidirectionalStream{
			StreamMessage: &BidirectionalStream_EdrCommand{
				EdrCommand: &EdrCommand{
					AgentId:    agentIDStr,
					CmdId:      cmdID,
					Action:     action,
					Payload:    payload,
					ExecutedBy: executedBy,
					Reason:     reason,
				},
			},
		})
	}()

	return &DispatchEdrCommandResponse{Dispatched: true, CmdId: cmdID}, nil
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

	// Replace any prior entry rather than rejecting the reconnect. A dead
	// prior stream's goroutine may still be looping on Recv (see
	// utils.WaitForReconnect) and would otherwise block the agent from
	// re-registering for minutes. evictIfOwner guards the map so the old
	// goroutine's eventual delete does not clobber the fresh entry.
	s.AgentStreamMutex.Lock()
	s.AgentStreamMap[idUint] = stream
	s.AgentStreamMutex.Unlock()

	defer func() {
		s.AgentStreamMutex.Lock()
		if s.AgentStreamMap[idUint] == stream {
			delete(s.AgentStreamMap, idUint)
		}
		s.AgentStreamMutex.Unlock()
	}()

	if OnAgentConnectHook != nil {
		OnAgentConnectHook(stream.Context(), idUint)
	}

	go serverHeartbeatLoop(stream.Context(), idUint, stream)

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			// The agent finished sending. It does not resume on this stream —
			// it opens a new one — so there is nothing here to wait for.
			return nil
		}
		if err != nil {
			catcher.Info("AgentStream: Recv errored, evicting stream",
				map[string]any{"agent_id": idUint, "err": err.Error(), "process": "agent-manager"})
			s.evictIfOwner(idUint, stream)
			return status.Error(codes.Internal, fmt.Sprintf("failed to receive message: %v", err))
		}

		switch msg := in.StreamMessage.(type) {
		case *BidirectionalStream_Heartbeat:
			if err := sendToAgent(idUint, stream, &BidirectionalStream{
				StreamMessage: &BidirectionalStream_Heartbeat{Heartbeat: &Heartbeat{}},
			}); err != nil {
				return status.Error(codes.Internal, fmt.Sprintf("failed to answer heartbeat: %v", err))
			}
		case *BidirectionalStream_Result:
			catcher.Info("Received command result from agent", map[string]any{"agent_id": msg.Result.AgentId, "result": msg.Result.Result, "process": "agent-manager"})
			if !s.tryDeliverResult(msg.Result) &&
				(OnCommandResultHook == nil || !OnCommandResultHook(msg.Result)) {
				catcher.Error("failed to find result channel for CmdID", nil, map[string]any{"cmdID": msg.Result.GetCmdId(), "process": "agent-manager"})
			}
		case *BidirectionalStream_ConfigState:
			for _, update := range s.diffConfig(idUint, msg.ConfigState.GetRevisions()) {
				if err := sendToAgent(idUint, stream, &BidirectionalStream{
					StreamMessage: &BidirectionalStream_ConfigUpdate{ConfigUpdate: update},
				}); err != nil {
					return status.Error(codes.Internal, fmt.Sprintf("failed to send config update: %v", err))
				}
			}
		case *BidirectionalStream_EdrStatusReport:
			// The authenticated context id is trustworthy; the agent_id
			// field inside the message is not.
			s.storeEdrStatus(idUint, msg.EdrStatusReport)
		case *BidirectionalStream_EdrResult:
			s.storeEdrResult(msg.EdrResult)
		}
	}
}

// serverHeartbeatLoop pings the agent on its own cadence, independent of the
// agent's own heartbeats or the reactive echo above. Two nginx L7 hops sit
// between them, and each one enforces its own idle timeout on its own side of
// the connection — a reply that only ever follows something the agent sent
// does not, by itself, guarantee the hop closest to this server ever sees
// server-initiated traffic.
func serverHeartbeatLoop(ctx context.Context, agentID uint, stream AgentService_AgentStreamServer) {
	t := time.NewTicker(agentHeartbeatInterval)
	defer t.Stop()
	msg := &BidirectionalStream{
		StreamMessage: &BidirectionalStream_Heartbeat{Heartbeat: &Heartbeat{}},
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			var lock sync.Locker
			if LockStreamHook != nil {
				lock = LockStreamHook(agentID)
			}
			var err error
			func() {
				if lock != nil {
					lock.Lock()
					defer lock.Unlock()
				}
				err = stream.Send(msg)
			}()
			if err != nil {
				return
			}
		}
	}
}

// tryDeliverResult hands an agent command result to the panel waiting on it.
// It must never block: the AgentStream goroutine is the single consumer of the
// agent's socket, and a stuck send here freezes every result for that agent
// (and the global CommandResultChannelM for the whole service). A missing
// slot, or a slot whose panel already went away (result unclaimed), returns
// false without blocking.
func (s *AgentService) tryDeliverResult(result *CommandResult) bool {
	cmdID := result.GetCmdId()
	s.CommandResultChannelM.Lock()
	defer s.CommandResultChannelM.Unlock()
	resultChan, ok := s.CommandResultChannel[cmdID]
	if !ok {
		return false
	}
	select {
	case resultChan <- result:
		return true
	default:
		// Slot is full: the panel that owned it left without claiming the
		// result. Drop the stale slot so the stream keeps flowing.
		delete(s.CommandResultChannel, cmdID)
		catcher.Error("dropping agent result: panel disconnected before claiming it", nil, map[string]any{"cmdID": cmdID, "process": "agent-manager"})
		return false
	}
}

// reclaimResultSlot removes the panel command slot, releasing any AgentStream
// goroutine blocked on delivering a result for it. Called via defer on every
// exit path of handlePanelCommand.
func (s *AgentService) reclaimResultSlot(cmdID string) {
	s.CommandResultChannelM.Lock()
	delete(s.CommandResultChannel, cmdID)
	s.CommandResultChannelM.Unlock()
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
		if err := s.handlePanelCommand(stream, cmd); err != nil {
			return err
		}
	}
}

// handlePanelCommand forwards one panel command to the agent stream and waits
// for the result (max 5 minutes). The result slot is created before the send
// and reclaimed via defer on EVERY exit — success, timeout, and panel
// disconnect — so a result that arrives after the panel is gone can never
// block the AgentStream goroutine.
func (s *AgentService) handlePanelCommand(stream PanelService_ProcessCommandServer, cmd *UtmCommand) error {
	streamId, err := strconv.Atoi(cmd.AgentId)
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid agent ID")
	}
	agentStream, ok := s.AgentStreamMap[uint(streamId)]
	if !ok {
		return status.Errorf(codes.NotFound, "agent not found or is disconnected")
	}

	target := &models.Agent{}
	if dErr := s.DBConnection.GetFirst(target, "id = ?", streamId); dErr == nil && target.NoRemoteControl {
		return status.Errorf(codes.PermissionDenied,
			"agent %d was installed with remote control disabled; it can only be changed on the machine itself", streamId)
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
	// Buffered by 1: a result can always be absorbed even when the panel is no
	// longer waiting, so delivery in AgentStream never blocks on it.
	s.CommandResultChannel[cmdID] = make(chan *CommandResult, 1)
	s.CommandResultChannelM.Unlock()
	defer s.reclaimResultSlot(cmdID)

	histCommand := createHistoryCommand(cmd, cmdID, uint(streamId))
	if cErr := s.DBConnection.Create(&histCommand); cErr != nil {
		catcher.Error("unable to create a new command history", cErr, map[string]any{"process": "agent-manager"})
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
		if uErr := s.DBConnection.Upsert(
			&models.AgentCommand{},
			"agent_id = ? AND cmd_id = ?",
			map[string]interface{}{"command_status": models.Executed, "result": result.Result},
			cmd.AgentId, cmdID,
		); uErr != nil {
			catcher.Error("failed to update command status", uErr, map[string]any{"process": "agent-manager"})
		}

		return stream.Send(result)
	case <-time.After(5 * time.Minute):
		_ = s.DBConnection.Upsert(
			&models.AgentCommand{},
			"agent_id = ? AND cmd_id = ?",
			map[string]interface{}{"command_status": models.Error, "result": "command timed out after 5 minutes"},
			cmd.AgentId, cmdID,
		)

		return status.Errorf(codes.DeadlineExceeded, "agent did not respond within 5 minutes")
	case <-stream.Context().Done():
		// The panel went away (viewer disconnected, backend request canceled).
		// Mark the history row, then exit without waiting: the deferred slot
		// reclamation keeps the AgentStream goroutine unblocked when the agent
		// finally responds.
		_ = s.DBConnection.Upsert(
			&models.AgentCommand{},
			"agent_id = ? AND cmd_id = ?",
			map[string]interface{}{"command_status": models.Error, "result": "panel disconnected before the agent responded"},
			cmd.AgentId, cmdID,
		)

		return stream.Context().Err()
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

func sendToAgent(agentID uint, stream AgentService_AgentStreamServer, msg *BidirectionalStream) error {
	var lock sync.Locker
	if LockStreamHook != nil {
		lock = LockStreamHook(agentID)
	}
	if lock != nil {
		lock.Lock()
		defer lock.Unlock()
	}
	return stream.Send(msg)
}
