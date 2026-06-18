package agent

import (
	context "context"
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
)

var (
	CollectorServ     *CollectorService
	collectorServOnce sync.Once
)

type ConfigStatus int32

const (
	ConfigSent    ConfigStatus = 1
	ConfigPending ConfigStatus = 2
)

type CollectorService struct {
	UnimplementedCollectorServiceServer

	CollectorStreamMap     map[uint]CollectorService_CollectorStreamServer
	CollectorStreamMutex   sync.Mutex
	CacheCollectorKey      map[uint]string
	CacheCollectorKeyMutex sync.RWMutex

	DBConnection *database.DB
}

func (s *CollectorService) ValidateCollectorKey(key string, id uint) bool {
	s.CacheCollectorKeyMutex.RLock()
	defer s.CacheCollectorKeyMutex.RUnlock()
	_, valid := utils.IsKeyPairValid(key, id, s.CacheCollectorKey)
	return valid
}

func InitCollectorService() {
	collectorServOnce.Do(func() {
		CollectorServ = &CollectorService{
			CollectorStreamMap: make(map[uint]CollectorService_CollectorStreamServer),
			CacheCollectorKey:  make(map[uint]string),
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
			CollectorServ.CacheCollectorKey[c.ID] = c.CollectorKey
		}
	})
}

func (s *CollectorService) RegisterCollector(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	collector := &models.Collector{
		Ip:       req.GetIp(),
		Hostname: req.GetHostname(),
		Version:  req.GetVersion(),
		Module:   models.CollectorModule(req.GetCollector().String()),
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
	s.CacheCollectorKey[collector.ID] = key
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
	s.CacheCollectorKeyMutex.Unlock()

	s.CollectorStreamMutex.Lock()
	delete(s.CollectorStreamMap, uint(idInt))
	s.CollectorStreamMutex.Unlock()

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

		case *CollectorMessages_Config:
			// Not implemented
		}
	}
}
