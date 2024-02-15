package agent

import (
	"context"
	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

func (s *Grpc) RegisterCollector(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
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
	res := &RegisterResponse{
		Id:           uint32(collector.ID),
		CollectorKey: key,
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
