package agent

import (
	"context"
	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/agent-manager/models"
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
