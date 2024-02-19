package agent

import (
	"context"
	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"time"
)

func (s *Grpc) RegisterCollector(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	collector := &models.Collector{
		Ip:       req.GetIp(),
		Hostname: req.GetHostname(),
		Version:  req.GetVersion(),
		Module:   models.CollectorModule(req.GetCollector().String()),
	}
	key := uuid.New().String()
	collector.CollectorKey = key
	err := collectorService.Create(collector)
	if err != nil {
		return nil, err
	}
	s.cacheMutex.Lock()
	CacheCollector[collector.ID] = key
	s.cacheMutex.Unlock()
	err = lastSeenService.Set(key, time.Now())
	if err != nil {
		return nil, err
	}
	res := &AuthResponse{
		Id:  uint32(collector.ID),
		Key: key,
	}
	return res, nil
}

func (s *Grpc) LoadCollectorsCacheFromDatabase() error {
	collectors, err := collectorService.FindAll()
	if err != nil {
		log.Fatalf("Failed to fetch collectors from database: %v", err)
		return err
	}
	for _, agent := range collectors {
		CacheCollector[agent.ID] = agent.CollectorKey
	}
	return nil
}

func (s *Grpc) ListCollector(ctx context.Context, req *ListRequest) (*ListCollectorResponse, error) {
	page := util.NewPaginator(int(req.PageSize), int(req.PageNumber), req.SortBy)

	filter := util.NewFilter(req.SearchQuery)

	collectors, total, err := collectorService.ListCollectors(page, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch agents: %v", err)
	}
	return convertToCollectorResponse(collectors, total)
}

func convertToCollectorResponse(agents []models.Collector, total int64) (*ListCollectorResponse, error) {
	var collectorMessages []*Collector
	for _, collector := range agents {
		agentProto := parseCollectorToProto(collector)
		collectorMessages = append(collectorMessages, agentProto)
	}
	return &ListCollectorResponse{
		Rows:  collectorMessages,
		Total: int32(total),
	}, nil
}

func parseCollectorToProto(collector models.Collector) *Collector {
	collectorStatus, lastSeen := lastSeenService.GetStatus(collector.CollectorKey)
	return &Collector{
		Id:           int32(collector.ID),
		Ip:           collector.Ip,
		Status:       Status(collectorStatus),
		Hostname:     collector.Hostname,
		Version:      collector.Version,
		LastSeen:     lastSeen,
		CollectorKey: collector.CollectorKey,
	}
}

func (s *Grpc) CollectorConfigStream(stream CollectorConfigurationService_CollectorConfigStreamServer) error {
	for {
		collectorConfig, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// Get the agent from the agents map
		collectorStream, ok := s.CollectorStreamMap[collectorConfig.CollectorKey]
		if !ok {
			return status.Errorf(codes.NotFound, "agent not found or is disconnected")
		}

		chanId := uuid.New().String()
		// Create a channel for the agent result and store it in the map
		s.configChannelM.Lock()
		resultChan := make(chan *ConfigKnowledge)
		s.ConfigChannel[chanId] = resultChan
		s.configChannelM.Unlock()

		// Send the command to the agent along with the command ID
		err = collectorStream.stream.Send(&CollectorMessages{
			StreamMessage: &CollectorMessages_Config{
				Config: &CollectorConfig{
					CollectorKey: collectorConfig.CollectorKey,
					Groups:       collectorConfig.Groups,
					RequestId:    chanId,
				},
			},
		})
		if err != nil {
			return err
		}

		result := <-resultChan
		err = stream.Send(result)
		if err != nil {
			return err
		}

		s.configChannelM.Lock()
		delete(s.ConfigChannel, chanId)
		s.configChannelM.Unlock()
	}
}

func (s *Grpc) CollectorStream(stream CollectorService_CollectorStreamServer) error {
	collectorKey, err := s.authenticateCollector(stream)
	if err != nil {
		return err
	}

	s.configMutex.Lock()
	s.CollectorStreamMap[collectorKey] = &StreamCollector{key: collectorKey, stream: stream}
	s.configMutex.Unlock()

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			s.configMutex.Lock()
			delete(s.CollectorStreamMap, collectorKey)
			s.configMutex.Unlock()
			collectorKey, err = s.authenticateCollector(stream)
			if err != nil {
				return err
			}
			s.configMutex.Lock()
			s.CollectorStreamMap[collectorKey] = &StreamCollector{key: collectorKey, stream: stream}
			s.configMutex.Unlock()
			return nil
		}
		if err != nil {
			return err
		}

		switch msg := in.StreamMessage.(type) {
		case *CollectorMessages_Config:
			if msg.Config.GetInternalKey() != config.GetInternalKey() {
				log.Printf("unauthorized config change attempt detected")
				continue
			}

		case *CollectorMessages_Result:
			log.Printf("Received Knowlodge: %s", msg.Result.RequestId)
			chanId := msg.Result.RequestId

			// Send the result back to the server
			if err := stream.Send(&CollectorMessages{
				StreamMessage: &CollectorMessages_Result{
					Result: &ConfigKnowledge{
						Accepted:  msg.Result.Accepted,
						RequestId: chanId,
					},
				},
			}); err != nil {
				log.Printf("Failed to send result to server: %v", err)
			}
			s.configChannelM.Lock()
			if resultChan, ok := s.ConfigChannel[chanId]; ok {
				resultChan <- &ConfigKnowledge{
					Accepted:  msg.Result.Accepted,
					RequestId: chanId,
				}

			} else {
				log.Printf("Failed to change configuration for request: %s", chanId)
			}
			s.configChannelM.Unlock()
		}
	}
}
