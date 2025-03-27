package agent

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *Grpc) RegisterCollector(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	collector := &models.Collector{
		Ip:       req.GetIp(),
		Hostname: req.GetHostname(),
		Version:  req.GetVersion(),
		Module:   models.CollectorModule(req.GetCollector().String()),
	}

	oldCollector, err := collectorService.GetCollectorByHostnameAndModule(collector.Hostname, string(collector.Module))
	if err == nil && len(oldCollector) > 0 {
		if oldCollector[0].Ip == collector.Ip {
			return &AuthResponse{
				Id:  uint32(oldCollector[0].ID),
				Key: oldCollector[0].CollectorKey,
			}, nil
		} else {
			util.Logger.ErrorF("Collector %s(%s) with id %d already registered with different IP", oldCollector[0].Hostname, oldCollector[0].Module, oldCollector[0].ID)
			return nil, status.Errorf(codes.AlreadyExists, "hostname has already been registered")
		}
	}

	key := uuid.New().String()
	collector.CollectorKey = key
	err = collectorService.Create(collector)
	if err != nil {
		util.Logger.ErrorF("Failed to create collector: %v", err)
		return nil, err
	}

	s.cacheCollectorMutex.Lock()
	CacheCollector[collector.ID] = key
	s.cacheCollectorMutex.Unlock()

	err = lastSeenService.Set(key, time.Now())
	if err != nil {
		util.Logger.ErrorF("Failed to set last seen: %v", err)
		return nil, err
	}
	res := &AuthResponse{
		Id:  uint32(collector.ID),
		Key: key,
	}

	util.Logger.Info("Collector %s(%s) with id %d registered correctly", collector.Hostname, collector.Module, collector.ID)
	return res, nil
}

func (s *Grpc) DeleteCollector(ctx context.Context, req *CollectorDelete) (*AuthResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "unable to get metadata from context")
	}

	keys, ok := md["key"]
	if !ok || len(keys) == 0 {
		return nil, status.Error(codes.Internal, "unable to get key from metadata")
	}
	key := keys[0]

	id, err := collectorService.Delete(uuid.MustParse(key), req.DeletedBy)
	if err != nil {
		util.Logger.ErrorF("unable to delete collector: %v", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to delete collector: %v", err.Error()))
	}

	s.cacheCollectorMutex.Lock()
	delete(CacheCollector, id)
	s.cacheCollectorMutex.Unlock()

	s.collectorStreamMutex.Lock()
	delete(s.CollectorStreamMap, key)
	s.collectorStreamMutex.Unlock()

	util.Logger.Info("Collector with key %s deleted by %s", key, req.DeletedBy)

	return &AuthResponse{
		Id:  uint32(id),
		Key: key,
	}, nil
}

func (s *Grpc) ListCollector(ctx context.Context, req *ListRequest) (*ListCollectorResponse, error) {
	page := util.NewPaginator(int(req.PageSize), int(req.PageNumber), req.SortBy)

	filter := util.NewFilter(req.SearchQuery)

	collectors, total, err := collectorService.ListCollectors(page, filter)
	if err != nil {
		util.Logger.ErrorF("failed to fetch collectors: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch collectors: %v", err)
	}
	return convertToCollectorResponse(collectors, total)
}

func (s *Grpc) ProcessPendingConfigs() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			s.pendingConfigM.Lock()
			copyPendingConfigs := make(map[string]string)
			for key, value := range s.PendingConfigs {
				copyPendingConfigs[key] = value
			}
			s.pendingConfigM.Unlock()

			for key := range copyPendingConfigs {

				s.collectorStreamMutex.Lock()
				stream, ok := s.CollectorStreamMap[key]
				s.collectorStreamMutex.Unlock()

				if ok {
					collector, err := collectorService.GetByKey(key)
					if err != nil {
						util.Logger.ErrorF("unable to get collector config to send config to stream : %v", err)
						continue
					}

					err = stream.Send(&CollectorMessages{
						StreamMessage: &CollectorMessages_Config{
							Config: convertToCollectorConfig(collector),
						},
					})
					if err != nil {
						util.Logger.ErrorF("failed to send config to collector: %v", err)
					}
				}
			}
		}
	}()
}

func (s *Grpc) CollectorStream(stream CollectorService_CollectorStreamServer) error {
	collectorKey, err := s.authenticateConnector(stream, ConnectorType_COLLECTOR)
	if err != nil {
		return err
	}

	s.collectorStreamMutex.Lock()
	if _, ok := s.CollectorStreamMap[collectorKey]; ok {
		s.collectorStreamMutex.Unlock()
		return fmt.Errorf("client %s is already connected", collectorKey)
	}
	s.CollectorStreamMap[collectorKey] = stream
	s.collectorStreamMutex.Unlock()

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			err = s.waitForReconnect(s.CollectorStreamMap[collectorKey].Context(), collectorKey, ConnectorType_COLLECTOR)
			if err != nil {
				s.collectorStreamMutex.Lock()
				delete(s.CollectorStreamMap, collectorKey)
				s.collectorStreamMutex.Unlock()

				util.Logger.ErrorF("failed to reconnect to client: %v", err)
				return fmt.Errorf("failed to reconnect to client: %v", err)
			}

			collectorKey, err = s.authenticateConnector(stream, ConnectorType_COLLECTOR)
			if err != nil {
				s.collectorStreamMutex.Lock()
				delete(s.CollectorStreamMap, collectorKey)
				s.collectorStreamMutex.Unlock()

				return err
			}

			continue
		}
		if err != nil {
			s.collectorStreamMutex.Lock()
			delete(s.CollectorStreamMap, collectorKey)
			s.collectorStreamMutex.Unlock()
			return err
		}

		switch msg := in.StreamMessage.(type) {
		case *CollectorMessages_Result:
			util.Logger.Info("Received Knowlodge: %s", msg.Result.RequestId)

			s.pendingConfigM.Lock()
			if s.PendingConfigs[collectorKey] == msg.Result.RequestId {
				delete(s.PendingConfigs, collectorKey)
			}
			s.pendingConfigM.Unlock()

		case *CollectorMessages_Config:
			// Not implemented
		}
	}
}

func (s *Grpc) GetCollectorConfig(ctx context.Context, in *ConfigRequest) (*CollectorConfig, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "unable to get metadata from context")
	}

	keys, ok := md["key"]
	if !ok || len(keys) == 0 {
		return nil, status.Error(codes.Internal, "unable to get key from metadata")
	}
	key := keys[0]

	collector, err := collectorService.GetByKey(key)
	if err != nil {
		util.Logger.ErrorF("unable to get collector config: %v", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to get collector config: %v", err.Error()))
	}

	return convertToCollectorConfig(collector), nil
}

func (s *Grpc) RegisterCollectorConfig(ctx context.Context, in *CollectorConfig) (*ConfigKnowledge, error) {
	collectorConfig := &CollectorConfig{
		CollectorKey: in.GetCollectorKey(),
		Groups:       in.GetGroups(),
		RequestId:    in.GetRequestId(),
	}

	collectorKey := collectorConfig.CollectorKey
	if collectorKey == "" {
		return nil, status.Errorf(codes.NotFound, "collector key is not provided")
	}

	collector, err := collectorService.GetByKey(collectorKey)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "collector not found in database or is deleted")
	}

	collectorConf := protoToModelCollectorGroups(collectorConfig.Groups, collector.ID)

	err = collectorService.SaveCollectorConfigs(collectorConf, collector.ID)
	if err != nil {
		util.Logger.ErrorF("error saving collector configuration: %v", err)
		return nil, status.Errorf(codes.Internal, "error saving collector configuration: %v", err.Error())
	}

	s.pendingConfigM.Lock()
	s.PendingConfigs[collectorKey] = collectorConfig.RequestId
	s.pendingConfigM.Unlock()

	return &ConfigKnowledge{
		Accepted:  "true",
		RequestId: collectorConfig.RequestId,
	}, nil
}

func (s *Grpc) ListCollectorHostnames(ctx context.Context, req *ListRequest) (*CollectorHostnames, error) {
	page := util.NewPaginator(int(req.PageSize), int(req.PageNumber), req.SortBy)
	filter := util.NewFilter(req.SearchQuery)

	hostnames, _, err := collectorService.GetHostnames(page, filter)
	if err != nil {
		util.Logger.ErrorF("failed to fetch hostnames: %v", err)
		return nil, status.Errorf(codes.NotFound, "failed to fetch hostnames: %v", err)
	}

	return &CollectorHostnames{
		Hostname: hostnames,
	}, nil
}

func (s *Grpc) GetCollectorsByHostnameAndModule(ctx context.Context, filter *FilterByHostAndModule) (*ListCollectorResponse, error) {
	collectors, err := collectorService.GetCollectorByHostnameAndModule(filter.GetHostname(), filter.GetModule().String())
	if err != nil {
		util.Logger.ErrorF("unable to get hostname: %v", err)
		return nil, status.Errorf(codes.NotFound, "unable to get hostname: %v", err)
	}

	return convertToCollectorResponse(collectors, int64(len(collectors)))
}

func (s *Grpc) LoadCollectorsCacheFromDatabase() error {
	collectors, err := collectorService.FindAll()
	if err != nil {
		util.Logger.ErrorF("Failed to fetch collectors from database: %v", err)
		return err
	}
	for _, colect := range collectors {
		CacheCollector[colect.ID] = colect.CollectorKey
	}
	return nil
}

func convertToCollectorResponse(collectors []models.Collector, total int64) (*ListCollectorResponse, error) {
	var collectorMessages []*Collector
	for _, collector := range collectors {
		collectorProto := modelToProtoCollector(collector)
		collectorMessages = append(collectorMessages, collectorProto)
	}
	return &ListCollectorResponse{
		Rows:  collectorMessages,
		Total: int32(total),
	}, nil
}

func convertToCollectorConfig(collectorConfig *models.Collector) *CollectorConfig {
	var protoGroups []*CollectorConfigGroup
	for _, group := range collectorConfig.Groups {
		protoGroup := &CollectorConfigGroup{
			Id:               int32(group.ID),
			GroupName:        group.GroupName,
			GroupDescription: group.GroupDescription,
			CollectorId:      int32(group.CollectorID),
		}

		for _, config := range group.Configurations {
			protoConfig := &CollectorGroupConfigurations{
				GroupId:         int32(config.ConfigGroupID),
				ConfKey:         config.ConfKey,
				ConfValue:       config.ConfValue,
				ConfName:        config.ConfName,
				ConfDescription: config.ConfDescription,
				ConfDataType:    config.ConfDataType,
				ConfRequired:    config.ConfRequired,
			}
			protoGroup.Configurations = append(protoGroup.Configurations, protoConfig)
		}

		protoGroups = append(protoGroups, protoGroup)
	}
	cnf := &CollectorConfig{
		CollectorKey: collectorConfig.CollectorKey,
		RequestId:    "",
		Groups:       protoGroups,
	}
	return cnf
}
