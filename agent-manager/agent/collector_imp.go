package agent

import (
	context "context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	sync "sync"
	"time"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/agent-manager/database"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/utils"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	"gorm.io/gorm"
)

var (
	CollectorServ     *CollectorService
	collectorServOnce sync.Once
)

var configAckTimeout = 15 * time.Second

type ConfigStatus int32

const (
	ConfigSent    ConfigStatus = 1
	ConfigPending ConfigStatus = 2
	ConfigAcked   ConfigStatus = 3
	ConfigFailed  ConfigStatus = 4
)

func (c ConfigStatus) String() string {
	switch c {
	case ConfigSent:
		return "Sent"
	case ConfigPending:
		return "Pending"
	case ConfigAcked:
		return "Acked"
	case ConfigFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

type CollectorService struct {
	UnimplementedCollectorServiceServer

	CollectorStreamMap     map[uint]CollectorService_CollectorStreamServer
	CollectorStreamMutex   sync.Mutex
	CacheCollectorKey      map[uint]utils.ConnectorAuth
	CacheCollectorKeyMutex sync.RWMutex
	CollectorSendMutex     sync.Map
	PendingAcks            map[string]*pendingAck
	PendingAcksMutex       sync.Mutex

	DBConnection *database.DB
}

type pendingAck struct {
	done            chan struct{}
	result          *ConfigKnowledge
	skipPersistence bool
}

func (s *CollectorService) registerOrJoinAck(requestID string, skipPersistence bool) (ack *pendingAck, joined bool) {
	s.PendingAcksMutex.Lock()
	defer s.PendingAcksMutex.Unlock()
	if existing, ok := s.PendingAcks[requestID]; ok {
		return existing, true
	}
	ack = &pendingAck{done: make(chan struct{}), skipPersistence: skipPersistence}
	s.PendingAcks[requestID] = ack
	return ack, false
}

func (s *CollectorService) pendingAckSkipsPersistence(requestID string) bool {
	s.PendingAcksMutex.Lock()
	defer s.PendingAcksMutex.Unlock()
	ack, ok := s.PendingAcks[requestID]
	return ok && ack.skipPersistence
}

func (s *CollectorService) finishAck(requestID string, ack *pendingAck, result *ConfigKnowledge) {
	s.PendingAcksMutex.Lock()
	cur, ok := s.PendingAcks[requestID]
	owns := ok && cur == ack
	if owns {
		delete(s.PendingAcks, requestID)
	}
	s.PendingAcksMutex.Unlock()
	if !owns {
		return
	}
	ack.result = result
	close(ack.done)
}

func (s *CollectorService) unregisterAckIfOwner(requestID string, ack *pendingAck) {
	s.PendingAcksMutex.Lock()
	if cur, ok := s.PendingAcks[requestID]; ok && cur == ack {
		delete(s.PendingAcks, requestID)
	}
	s.PendingAcksMutex.Unlock()
}

func (s *CollectorService) ValidateCollectorKey(key string, id uint) bool {
	s.CacheCollectorKeyMutex.RLock()
	defer s.CacheCollectorKeyMutex.RUnlock()
	_, valid := utils.IsKeyPairValid(key, id, s.CacheCollectorKey)
	return valid
}

func (s *CollectorService) sendLockFor(collectorID uint) *sync.Mutex {
	mu := &sync.Mutex{}
	actual, _ := s.CollectorSendMutex.LoadOrStore(collectorID, mu)
	return actual.(*sync.Mutex)
}

func InitCollectorService() {
	collectorServOnce.Do(func() {
		CollectorServ = &CollectorService{
			CollectorStreamMap: make(map[uint]CollectorService_CollectorStreamServer),
			CacheCollectorKey:  make(map[uint]utils.ConnectorAuth),
			PendingAcks:        make(map[string]*pendingAck),
			DBConnection:       database.GetDB(),
		}
		collectors := []models.Collector{}
		_, err := CollectorServ.DBConnection.GetAll(&collectors, "")
		if err != nil {
			_ = catcher.Error("failed to fetch collectors", err, map[string]any{"process": "agent-manager"})
			time.Sleep(5 * time.Second)
			os.Exit(1)
		}
		for _, c := range collectors {
			CollectorServ.CacheCollectorKey[c.ID] = utils.ConnectorAuth{Key: c.CollectorKey, TenantID: tenantOrDefault(c.TenantID)}
		}
	})
}

func (s *CollectorService) RegisterCollector(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	tenantID, ok := tenantFromContext(ctx)
	if !ok {
		// The interceptor is the only writer of this key on the connection-key
		// path; missing it means the route was reached through some other auth
		// that shouldn't be allowed to enrol collectors.
		return nil, status.Error(codes.PermissionDenied, "missing tenant on register")
	}

	collector := &models.Collector{
		Ip:       req.GetIp(),
		Hostname: req.GetHostname(),
		Version:  req.GetVersion(),
		Module:   models.CollectorModule(req.GetCollector().String()),
		TenantID: tenantID,
	}

	oldCollector := &models.Collector{}
	err := s.DBConnection.GetFirst(oldCollector, "hostname = ? and module = ?", collector.Hostname, string(collector.Module))
	if err == nil {
		if oldCollector.Ip == collector.Ip {
			return &AuthResponse{
				Id:  uint32(oldCollector.ID),
				Key: oldCollector.CollectorKey,
			}, nil
		} else {
			catcher.Error("collector already registered with different IP", nil, map[string]any{"hostname": oldCollector.Hostname, "module": oldCollector.Module, "id": oldCollector.ID, "process": "agent-manager"})
			return nil, status.Errorf(codes.AlreadyExists, "hostname has already been registered")
		}
	}

	key := uuid.New().String()
	collector.CollectorKey = key
	err = s.DBConnection.Create(collector)
	if err != nil {
		catcher.Error("failed to create collector", err, map[string]any{"process": "agent-manager"})
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create collector: %v", err))
	}

	s.CacheCollectorKeyMutex.Lock()
	entry := utils.ConnectorAuth{Key: key, TenantID: collector.TenantID}
	s.CacheCollectorKey[collector.ID] = entry
	AuthCache.PublishCollector(collector.ID, entry)
	s.CacheCollectorKeyMutex.Unlock()

	LastSeenChannel <- models.LastSeen{
		ConnectorType: "collector",
		ConnectorID:   collector.ID,
		LastPing:      time.Now(),
	}

	catcher.Info("Collector registered correctly", map[string]any{"hostname": collector.Hostname, "module": collector.Module, "id": collector.ID, "process": "agent-manager"})
	return &AuthResponse{
		Id:  uint32(collector.ID),
		Key: key,
	}, nil
}

func (s *CollectorService) DeleteCollector(ctx context.Context, req *DeleteRequest) (*AuthResponse, error) {
	id, key, _, err := utils.GetItemsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	err = s.DBConnection.Upsert(&models.Collector{}, "id = ?", map[string]interface{}{"deleted_by": req.DeletedBy}, id)
	if err != nil {
		catcher.Error("unable to delete collector", err, map[string]any{"process": "agent-manager"})
	}

	err = s.DBConnection.Delete(&models.Collector{}, "id = ?", false, id)
	if err != nil {
		catcher.Error("unable to delete collector", err, map[string]any{"process": "agent-manager"})
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to delete collector: %v", err.Error()))
	}

	s.CacheCollectorKeyMutex.Lock()
	delete(s.CacheCollectorKey, uint(idInt))
	AuthCache.DeleteCollector(uint(idInt))
	s.CacheCollectorKeyMutex.Unlock()

	s.CollectorStreamMutex.Lock()
	delete(s.CollectorStreamMap, uint(idInt))
	s.CollectorStreamMutex.Unlock()

	s.CollectorSendMutex.Delete(uint(idInt))

	catcher.Info("Collector deleted", map[string]any{"key": key, "deleted_by": req.DeletedBy, "process": "agent-manager"})
	return &AuthResponse{
		Id:  uint32(idInt),
		Key: key,
	}, nil
}

func (s *CollectorService) ListCollector(ctx context.Context, req *ListRequest) (*ListCollectorResponse, error) {
	page := utils.NewPaginator(int(req.PageSize), int(req.PageNumber), req.SortBy)
	filter := utils.NewFilter(req.SearchQuery)

	collectors := []models.Collector{}
	total, err := s.DBConnection.GetByPagination(&collectors, page, filter, "", false)
	if err != nil {
		catcher.Error("failed to fetch collectors", err, map[string]any{"process": "agent-manager"})
		return nil, status.Errorf(codes.Internal, "failed to fetch collectors: %v", err)
	}
	return convertModelToCollectorResponse(collectors, total), nil
}

func (s *CollectorService) CollectorStream(stream CollectorService_CollectorStreamServer) error {
	id, _, _, err := utils.GetItemsFromContext(stream.Context())
	if err != nil {
		return status.Error(codes.InvalidArgument, fmt.Errorf("unable to get items from context: %v", err).Error())
	}
	uid, err := strconv.Atoi(id)
	if err != nil {
		return status.Error(codes.InvalidArgument, fmt.Errorf("invalid id: %v", err).Error())
	}

	s.CollectorStreamMutex.Lock()
	if _, ok := s.CollectorStreamMap[uint(uid)]; ok {
		s.CollectorStreamMutex.Unlock()
		return status.Error(codes.AlreadyExists, "client is already connected")
	}
	s.CollectorStreamMap[uint(uid)] = stream
	s.CollectorStreamMutex.Unlock()

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			err = utils.WaitForReconnect(stream.Context(), stream)
			if err != nil {
				s.CollectorStreamMutex.Lock()
				delete(s.CollectorStreamMap, uint(uid))
				s.CollectorStreamMutex.Unlock()
				return status.Error(codes.Internal, fmt.Sprintf("failed to reconnect to client: %v", err))
			}
			continue
		}
		if err != nil {
			s.CollectorStreamMutex.Lock()
			delete(s.CollectorStreamMap, uint(uid))
			s.CollectorStreamMutex.Unlock()
			return status.Error(codes.Internal, fmt.Sprintf("failed to receive message from client: %v", err))
		}

		switch msg := in.StreamMessage.(type) {
		case *CollectorMessages_Result:
			catcher.Info("Received Knowledge", map[string]any{"request_id": msg.Result.RequestId, "process": "agent-manager"})
			s.handleConfigResult(msg.Result)

		case *CollectorMessages_Config:
			catcher.Warn("received unexpected Config message on collector stream (ignored)", map[string]any{"process": "agent-manager"})
		}
	}
}

func (s *CollectorService) handleConfigResult(result *ConfigKnowledge) {
	requestID := result.GetRequestId()
	if requestID == "" {
		return
	}

	if !s.pendingAckSkipsPersistence(requestID) {
		newStatus := ConfigAcked.String()
		lastError := ""
		if result.GetAccepted() != "true" {
			newStatus = ConfigFailed.String()
			lastError = result.GetErrorMessage()
		}

		updated, uerr := s.DBConnection.UpdateOnly(&models.CollectorIntegrationConfig{}, "request_id = ?",
			map[string]interface{}{"config_status": newStatus, "last_error": lastError}, requestID)
		if uerr != nil {
			catcher.Error("failed to persist collector config result", uerr, map[string]any{"request_id": requestID, "process": "agent-manager"})
		} else if !updated {
			catcher.Warn("discarding stale collector config ack: no row currently correlates to this request_id, superseded by a newer request on the same collector/data_type", map[string]any{"request_id": requestID, "process": "agent-manager"})
		}
	}

	s.resolveAckByID(requestID, result)
}

func (s *CollectorService) resolveAckByID(requestID string, result *ConfigKnowledge) bool {
	s.PendingAcksMutex.Lock()
	ack, ok := s.PendingAcks[requestID]
	if ok {
		delete(s.PendingAcks, requestID)
	}
	s.PendingAcksMutex.Unlock()
	if !ok {
		return false
	}
	ack.result = result
	close(ack.done)
	return true
}

const ReservedTLSCertsGroup = "__tls_certs__"

const (
	tlsCertConfKeyAction  = "action"
	tlsCertConfKeyCertPem = "certPem"
	tlsCertConfKeyKeyPem  = "keyPem"
	tlsCertConfKeyCaPem   = "caPem"
	tlsCertConfKeyPayload = "payload_enc"
	tlsCertConfKeyNonce   = "nonce"

	tlsCertActionApply  = "apply"
	tlsCertActionStatus = "status"
)

type tlsCertEnvelope struct {
	CertPem string `json:"certPem"`
	KeyPem  string `json:"keyPem"`
	CaPem   string `json:"caPem"`
}

func confValue(configs []*CollectorGroupConfigurations, key string) string {
	for _, c := range configs {
		if c.GetConfKey() == key {
			return c.GetConfValue()
		}
	}
	return ""
}

func (s *CollectorService) handleTLSCertsGroup(collectorID int, requestID string, group *CollectorConfigGroup) (*ConfigKnowledge, error) {
	s.CollectorStreamMutex.Lock()
	_, online := s.CollectorStreamMap[uint(collectorID)]
	s.CollectorStreamMutex.Unlock()
	if !online {
		return nil, status.Errorf(codes.Unavailable, "collector %d is offline", collectorID)
	}

	action := confValue(group.GetConfigurations(), tlsCertConfKeyAction)
	switch action {
	case tlsCertActionApply:
		return s.pushTLSCerts(collectorID, requestID, group)
	case tlsCertActionStatus:
		return s.queryTLSStatus(collectorID, requestID)
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown %s action %q", ReservedTLSCertsGroup, action)
	}
}

func (s *CollectorService) pushTLSCerts(collectorID int, requestID string, group *CollectorConfigGroup) (*ConfigKnowledge, error) {
	s.CacheCollectorKeyMutex.RLock()
	entry, known := s.CacheCollectorKey[uint(collectorID)]
	s.CacheCollectorKeyMutex.RUnlock()
	if !known {
		return nil, status.Errorf(codes.NotFound, "collector %d has no cached collector key", collectorID)
	}

	envelope := tlsCertEnvelope{
		CertPem: confValue(group.GetConfigurations(), tlsCertConfKeyCertPem),
		KeyPem:  confValue(group.GetConfigurations(), tlsCertConfKeyKeyPem),
		CaPem:   confValue(group.GetConfigurations(), tlsCertConfKeyCaPem),
	}
	plaintext, merr := json.Marshal(envelope)
	if merr != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal tls cert envelope: %v", merr)
	}

	key := utils.DeriveTLSCertKey(strconv.Itoa(collectorID), entry.Key)
	ciphertextB64, nonceB64, serr := utils.SealTLSCertEnvelope(key, plaintext)
	if serr != nil {
		return nil, status.Errorf(codes.Internal, "failed to seal tls cert envelope: %v", serr)
	}

	return s.relayReservedGroup(collectorID, requestID, []*CollectorGroupConfigurations{
		{ConfKey: tlsCertConfKeyAction, ConfValue: tlsCertActionApply},
		{ConfKey: tlsCertConfKeyPayload, ConfValue: ciphertextB64},
		{ConfKey: tlsCertConfKeyNonce, ConfValue: nonceB64},
	})
}

func (s *CollectorService) queryTLSStatus(collectorID int, requestID string) (*ConfigKnowledge, error) {
	return s.relayReservedGroup(collectorID, requestID, []*CollectorGroupConfigurations{
		{ConfKey: tlsCertConfKeyAction, ConfValue: tlsCertActionStatus},
	})
}

func (s *CollectorService) relayReservedGroup(collectorID int, requestID string, configs []*CollectorGroupConfigurations) (*ConfigKnowledge, error) {
	s.CollectorStreamMutex.Lock()
	stream, ok := s.CollectorStreamMap[uint(collectorID)]
	s.CollectorStreamMutex.Unlock()
	if !ok {
		return nil, status.Errorf(codes.Unavailable, "collector %d is offline", collectorID)
	}

	ack, joined := s.registerOrJoinAck(requestID, true)
	if joined {
		select {
		case <-ack.done:
			return ack.result, nil
		case <-time.After(configAckTimeout):
			return nil, status.Errorf(codes.DeadlineExceeded, "collector %d did not acknowledge %s within %s", collectorID, ReservedTLSCertsGroup, configAckTimeout)
		}
	}

	sendMu := s.sendLockFor(uint(collectorID))
	sendMu.Lock()
	sendErr := stream.Send(&CollectorMessages{
		StreamMessage: &CollectorMessages_Config{
			Config: &CollectorConfig{
				CollectorId: strconv.Itoa(collectorID),
				RequestId:   requestID,
				Groups: []*CollectorConfigGroup{
					{
						GroupName:      ReservedTLSCertsGroup,
						CollectorId:    int32(collectorID),
						Configurations: configs,
					},
				},
			},
		},
	})
	sendMu.Unlock()
	if sendErr != nil {
		s.finishAck(requestID, ack, &ConfigKnowledge{Accepted: "false", RequestId: requestID, ErrorMessage: sendErr.Error()})
		return nil, status.Errorf(codes.Internal, "failed to send %s to collector: %v", ReservedTLSCertsGroup, sendErr)
	}

	select {
	case <-ack.done:
		return ack.result, nil
	case <-time.After(configAckTimeout):
		s.unregisterAckIfOwner(requestID, ack)
		return nil, status.Errorf(codes.DeadlineExceeded, "collector %d did not acknowledge %s within %s", collectorID, ReservedTLSCertsGroup, configAckTimeout)
	}
}

func (s *CollectorService) GetCollectorConfig(ctx context.Context, req *ConfigRequest) (*CollectorConfig, error) {
	collectorID := req.GetCollectorId()
	if collectorID <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid collector id")
	}

	rows := []models.CollectorIntegrationConfig{}
	_, err := s.DBConnection.GetAll(&rows, "collector_id = ?", collectorID)
	if err != nil {
		catcher.Error("failed to fetch collector integration config", err, map[string]any{"collector_id": collectorID, "process": "agent-manager"})
		return nil, status.Errorf(codes.Internal, "failed to fetch collector config: %v", err)
	}

	groups := make([]*CollectorConfigGroup, 0, len(rows))
	for _, row := range rows {
		if row.DataType == ReservedTLSCertsGroup {
			continue
		}
		group := &CollectorConfigGroup{
			GroupName:   row.DataType,
			CollectorId: collectorID,
			RequestId:   row.RequestID,
		}
		if row.DesiredStateJSON != "" {
			var confs []*CollectorGroupConfigurations
			if uerr := json.Unmarshal([]byte(row.DesiredStateJSON), &confs); uerr != nil {
				catcher.Error("failed to unmarshal desired state", uerr, map[string]any{"collector_id": collectorID, "data_type": row.DataType, "process": "agent-manager"})
				continue
			}
			group.Configurations = confs
		}
		groups = append(groups, group)
	}

	return &CollectorConfig{
		CollectorId: strconv.Itoa(int(collectorID)),
		Groups:      groups,
	}, nil
}

func (s *CollectorService) SetCollectorConfig(ctx context.Context, req *CollectorConfig) (*ConfigKnowledge, error) {
	if len(req.GetGroups()) != 1 {
		return nil, status.Error(codes.InvalidArgument, "exactly one config group is required")
	}
	group := req.Groups[0]

	collectorID, err := strconv.Atoi(req.GetCollectorId())
	if err != nil || collectorID <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid collector id")
	}
	dataType := group.GetGroupName()
	if dataType == "" {
		return nil, status.Error(codes.InvalidArgument, "data type (group_name) is required")
	}

	requestID := req.GetRequestId()
	if requestID == "" {
		requestID = uuid.New().String()
	}

	if dataType == ReservedTLSCertsGroup {
		return s.handleTLSCertsGroup(collectorID, requestID, group)
	}

	existing := &models.CollectorIntegrationConfig{}
	if ferr := s.DBConnection.GetFirst(existing, "request_id = ?", requestID); ferr == nil {
		switch existing.ConfigStatus {
		case ConfigAcked.String():
			return &ConfigKnowledge{Accepted: "true", RequestId: requestID}, nil
		case ConfigFailed.String():
			return &ConfigKnowledge{Accepted: "false", RequestId: requestID, ErrorMessage: existing.LastError}, nil
		}
	}

	ack, joined := s.registerOrJoinAck(requestID, false)
	if joined {
		select {
		case <-ack.done:
			return ack.result, nil
		case <-time.After(configAckTimeout):
			return nil, status.Errorf(codes.DeadlineExceeded, "collector %d did not acknowledge config within %s", collectorID, configAckTimeout)
		}
	}

	configJSON, err := json.Marshal(group.GetConfigurations())
	if err != nil {
		s.unregisterAckIfOwner(requestID, ack)
		return nil, status.Errorf(codes.Internal, "failed to marshal desired state: %v", err)
	}

	row := &models.CollectorIntegrationConfig{
		CollectorID:      uint(collectorID),
		DataType:         dataType,
		DesiredStateJSON: string(configJSON),
		ConfigStatus:     ConfigPending.String(),
		RequestID:        requestID,
		LastError:        "",
	}
	if uerr := s.DBConnection.Upsert(row, "collector_id = ? AND data_type = ?", map[string]interface{}{
		"desired_state_json": row.DesiredStateJSON,
		"config_status":      row.ConfigStatus,
		"request_id":         row.RequestID,
		"last_error":         "",
	}, collectorID, dataType); uerr != nil {
		catcher.Error("failed to persist collector integration config", uerr, map[string]any{"collector_id": collectorID, "data_type": dataType, "process": "agent-manager"})
		s.unregisterAckIfOwner(requestID, ack)
		return nil, status.Errorf(codes.Internal, "failed to persist config: %v", uerr)
	}

	s.CollectorStreamMutex.Lock()
	stream, ok := s.CollectorStreamMap[uint(collectorID)]
	s.CollectorStreamMutex.Unlock()
	if !ok {
		s.unregisterAckIfOwner(requestID, ack)
		return nil, status.Errorf(codes.Unavailable, "collector %d is offline", collectorID)
	}

	group.CollectorId = int32(collectorID)
	sendMu := s.sendLockFor(uint(collectorID))
	sendMu.Lock()
	sendErr := stream.Send(&CollectorMessages{
		StreamMessage: &CollectorMessages_Config{
			Config: &CollectorConfig{
				CollectorId: req.GetCollectorId(),
				Groups:      []*CollectorConfigGroup{group},
				RequestId:   requestID,
			},
		},
	})
	sendMu.Unlock()
	if sendErr != nil {
		_, _ = s.DBConnection.UpdateOnly(&models.CollectorIntegrationConfig{}, "collector_id = ? AND data_type = ?",
			map[string]interface{}{"config_status": ConfigFailed.String(), "last_error": sendErr.Error()}, collectorID, dataType)
		s.finishAck(requestID, ack, &ConfigKnowledge{Accepted: "false", RequestId: requestID, ErrorMessage: sendErr.Error()})
		return nil, status.Errorf(codes.Internal, "failed to send config to collector: %v", sendErr)
	}

	_, _ = s.DBConnection.UpdateOnly(&models.CollectorIntegrationConfig{}, "collector_id = ? AND data_type = ?",
		map[string]interface{}{"config_status": ConfigSent.String()}, collectorID, dataType)

	select {
	case <-ack.done:
		return ack.result, nil
	case <-time.After(configAckTimeout):
		s.unregisterAckIfOwner(requestID, ack)
		return nil, status.Errorf(codes.DeadlineExceeded, "collector %d did not acknowledge config within %s", collectorID, configAckTimeout)
	}
}

func (s *CollectorService) GetCollectorIntegrationState(ctx context.Context, req *IntegrationStateRequest) (*IntegrationStateResponse, error) {
	collectorID := req.GetCollectorId()
	if collectorID <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid collector id")
	}
	dataType := req.GetDataType()
	if dataType == "" {
		return nil, status.Error(codes.InvalidArgument, "data type is required")
	}

	row := &models.CollectorIntegrationConfig{}
	if ferr := s.DBConnection.GetFirst(row, "collector_id = ? AND data_type = ?", collectorID, dataType); ferr != nil {
		if errors.Is(ferr, gorm.ErrRecordNotFound) {
			return &IntegrationStateResponse{Configured: false}, nil
		}
		catcher.Error("failed to fetch collector integration state", ferr, map[string]any{"collector_id": collectorID, "data_type": dataType, "process": "agent-manager"})
		return nil, status.Errorf(codes.Internal, "failed to fetch collector integration state: %v", ferr)
	}

	resp := &IntegrationStateResponse{
		Configured:   true,
		ConfigStatus: row.ConfigStatus,
		LastError:    row.LastError,
	}
	if row.DesiredStateJSON != "" {
		var confs []*CollectorGroupConfigurations
		if uerr := json.Unmarshal([]byte(row.DesiredStateJSON), &confs); uerr != nil {
			catcher.Error("failed to unmarshal desired state", uerr, map[string]any{"collector_id": collectorID, "data_type": dataType, "process": "agent-manager"})
			return nil, status.Errorf(codes.Internal, "failed to parse stored config: %v", uerr)
		}
		resp.Configurations = confs
	}
	return resp, nil
}
