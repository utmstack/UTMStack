package agent

import (
	context "context"
	"fmt"
	"io"
	"strconv"
	sync "sync"
	"time"

	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/database"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/utils"
	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
	"github.com/utmstack/config-client-go/types"
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
	UnimplementedPanelCollectorServiceServer

	CollectorStreamMap        map[uint]CollectorService_CollectorStreamServer
	CollectorStreamMutex      sync.Mutex
	CollectorConfigsCache     map[uint][]*CollectorConfigGroup
	CollectorConfigsCacheM    sync.Mutex
	CacheCollectorKey         map[uint]string
	CacheCollectorKeyMutex    sync.Mutex
	CollectorPendigConfigChan chan *CollectorConfig
	CollectorTypes            []enum.UTMModule

	DBConnection *database.DB
}

func InitCollectorService() error {
	var err error
	collectorServOnce.Do(func() {
		CollectorServ = &CollectorService{
			CollectorStreamMap:        make(map[uint]CollectorService_CollectorStreamServer),
			CollectorConfigsCache:     make(map[uint][]*CollectorConfigGroup),
			CacheCollectorKey:         make(map[uint]string),
			CollectorPendigConfigChan: make(chan *CollectorConfig, 1000),
			CollectorTypes:            []enum.UTMModule{enum.IBM_AS_400},
			DBConnection:              database.GetDB(),
		}
		collectors := []models.Collector{}
		_, err = CollectorServ.DBConnection.GetAll(&collectors, "")
		if err != nil {
			err = fmt.Errorf("failed to get collectors: %v", err)
			return
		}
		for _, c := range collectors {
			CollectorServ.CacheCollectorKey[c.ID] = c.CollectorKey
		}

		go CollectorServ.ProcessPendingConfigs()

		client := utmconf.NewUTMClient(config.GetInternalKey(), config.GetPanelServiceName())
		for _, moduleType := range CollectorServ.CollectorTypes {
			moduleConfig := &types.ConfigurationSection{}
			moduleConfig, err = client.GetUTMConfig(moduleType)
			if err != nil {
				err = fmt.Errorf("failed to get module config: %v", err)
				return
			}

			pendigConfigs := make(map[string][]*CollectorConfigGroup)
			for _, group := range moduleConfig.ConfigurationGroups {
				var idInt int
				idInt, err = strconv.Atoi(group.CollectorID)
				if err != nil {
					err = fmt.Errorf("failed to convert collector id to int: %v", err)
					return
				}

				CollectorServ.CollectorConfigsCache[uint(idInt)] = append(
					CollectorServ.CollectorConfigsCache[uint(idInt)],
					convertModuleGroupToCollectorProto(group),
				)

				pendigConfigs[group.CollectorID] = append(pendigConfigs[group.CollectorID], convertModuleGroupToCollectorProto(group))
			}

			for id, configs := range pendigConfigs {
				CollectorServ.CollectorPendigConfigChan <- &CollectorConfig{
					CollectorId: id,
					RequestId:   uuid.New().String(),
					Groups:      configs,
				}
			}
		}
	})
	return err
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
			utils.ALogger.ErrorF("collector %s(%s) with id %d already registered with different IP", oldCollector.Hostname, oldCollector.Module, oldCollector.ID)
			return nil, status.Errorf(codes.AlreadyExists, "hostname has already been registered")
		}
	}

	key := uuid.New().String()
	collector.CollectorKey = key
	err = s.DBConnection.Create(collector)
	if err != nil {
		utils.ALogger.ErrorF("failed to create collector: %v", err)
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

	utils.ALogger.Info("Collector %s(%s) with id %d registered correctly", collector.Hostname, collector.Module, collector.ID)
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
		utils.ALogger.ErrorF("unable to delete collector: %v", err)
	}

	err = s.DBConnection.Delete(&models.Collector{}, "id = ?", false, id)
	if err != nil {
		utils.ALogger.ErrorF("unable to delete collector: %v", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to delete collector: %v", err.Error()))
	}

	s.CacheCollectorKeyMutex.Lock()
	delete(s.CacheCollectorKey, uint(idInt))
	s.CacheCollectorKeyMutex.Unlock()

	s.CollectorStreamMutex.Lock()
	delete(s.CollectorStreamMap, uint(idInt))
	s.CollectorStreamMutex.Unlock()

	utils.ALogger.Info("Collector with key %s deleted by %s", key, req.DeletedBy)
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
		utils.ALogger.ErrorF("failed to fetch collectors: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch collectors: %v", err)
	}
	return convertModelToCollectorResponse(collectors, total), nil
}

func (s *CollectorService) ProcessPendingConfigs() {
	for configs := range s.CollectorPendigConfigChan {
		collectorID, err := strconv.Atoi(configs.CollectorId)
		if err != nil {
			utils.ALogger.ErrorF("invalid collector ID: %v", err)
			continue
		}

		s.CollectorStreamMutex.Lock()
		stream, ok := s.CollectorStreamMap[uint(collectorID)]
		s.CollectorStreamMutex.Unlock()

		if ok {
			err = stream.Send(&CollectorMessages{
				StreamMessage: &CollectorMessages_Config{
					Config: &CollectorConfig{
						Groups: configs.Groups,
					},
				},
			})
			if err != nil {
				utils.ALogger.ErrorF("failed to send config to collector: %v", err)
			}
		}
	}
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
			utils.ALogger.Info("Received Knowlodge: %s", msg.Result.RequestId)

		case *CollectorMessages_Config:
			// Not implemented
		}
	}
}

func (s *CollectorService) GetCollectorConfig(ctx context.Context, in *ConfigRequest) (*CollectorConfig, error) {
	id, _, _, err := utils.GetItemsFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("unable to get items from context: %v", err).Error())
	}
	uid, err := strconv.Atoi(id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("invalid id: %v", err).Error())
	}

	s.CollectorConfigsCacheM.Lock()
	defer s.CollectorConfigsCacheM.Unlock()

	return &CollectorConfig{
		Groups: s.CollectorConfigsCache[uint(uid)],
	}, nil
}

func (s *CollectorService) RegisterCollectorConfig(ctx context.Context, in *CollectorConfig) (*ConfigKnowledge, error) {
	collectorID, err := strconv.Atoi(in.CollectorId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid collector ID")
	}

	s.CollectorPendigConfigChan <- in

	s.CollectorConfigsCacheM.Lock()
	s.CollectorConfigsCache[uint(collectorID)] = in.Groups
	s.CollectorConfigsCacheM.Unlock()

	return &ConfigKnowledge{
		Accepted:  "true",
		RequestId: in.RequestId,
	}, nil
}
