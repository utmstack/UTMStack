package usecase

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/logstash/connectors"
	"github.com/utmstack/utmstack/backend/modules/logstash/domain"
	"github.com/utmstack/utmstack/backend/pkg/logger"
)

const (
	pipelineSchedulerInterval     = 20 * time.Second
	pipelineSchedulerInitialDelay = 30 * time.Second
)

type PipelineScheduler struct {
	pipelineRepo   connectors.PipelineRepository
	filterRepo     connectors.FilterRepository
	statisticsRepo connectors.StatisticsRepository
}

func NewPipelineScheduler(
	pipelineRepo connectors.PipelineRepository,
	filterRepo connectors.FilterRepository,
	statisticsRepo connectors.StatisticsRepository,
) *PipelineScheduler {
	return &PipelineScheduler{
		pipelineRepo:   pipelineRepo,
		filterRepo:     filterRepo,
		statisticsRepo: statisticsRepo,
	}
}

func (s *PipelineScheduler) Start(ctx context.Context) {
	logger.Info("pipeline scheduler: starting (initial delay 30s)")

	select {
	case <-time.After(pipelineSchedulerInitialDelay):
	case <-ctx.Done():
		logger.Info("pipeline scheduler: cancelled during initial delay — stopped")
		return
	}

	ticker := time.NewTicker(pipelineSchedulerInterval)
	defer ticker.Stop()

	logger.Info("pipeline scheduler: running")

	for {
		select {
		case <-ticker.C:
			s.tick(ctx)
		case <-ctx.Done():
			logger.Info("pipeline scheduler: context cancelled — stopped")
			return
		}
	}
}

func (s *PipelineScheduler) tick(ctx context.Context) {
	pipelines, err := s.pipelineRepo.AllActivePipelinesByServer(ctx)
	if err != nil {
		logger.Error("pipeline scheduler: AllActivePipelinesByServer error: " + err.Error())
		return
	}

	for i := range pipelines {
		p := &pipelines[i]
		dataType := s.resolveDataType(ctx, p)

		stat, err := s.statisticsRepo.GetLatestStatistic(ctx, dataType)
		if err != nil {
			logger.Error("pipeline scheduler: GetLatestStatistic error for pipeline " + safeStr(p.PipelineName) + ": " + err.Error())
			setStatus(p, pipelineStatusDown)
			continue
		}

		if stat == nil {
			setStatus(p, pipelineStatusDown)
			continue
		}

		// Parse timestamp and compute hours difference
		ts, parseErr := time.Parse(time.RFC3339, stat.Timestamp)
		if parseErr != nil {
			// Try alternative formats
			ts, parseErr = time.Parse("2006-01-02T15:04:05.000Z", stat.Timestamp)
			if parseErr != nil {
				ts, parseErr = time.Parse("2006-01-02T15:04:05Z", stat.Timestamp)
			}
		}
		if parseErr != nil {
			logger.Warn("pipeline scheduler: could not parse timestamp '" + stat.Timestamp + "': " + parseErr.Error())
			setStatus(p, pipelineStatusDown)
			continue
		}

		hoursDiff := time.Since(ts).Hours()
		if hoursDiff > 6 {
			setStatus(p, pipelineStatusDown)
		} else {
			setStatus(p, pipelineStatusUp)
		}
		p.EventsOut = &stat.Count
	}

	if err := s.pipelineRepo.SaveAll(ctx, pipelines); err != nil {
		logger.Error("pipeline scheduler: SaveAll error: " + err.Error())
	}
}

func (s *PipelineScheduler) resolveDataType(ctx context.Context, p *domain.UtmLogstashPipeline) string {
	fallback := safeStr(p.PipelineName)

	if p.ModuleName == nil || *p.ModuleName == "" {
		return fallback
	}

	filters, err := s.filterRepo.FindByModuleName(ctx, *p.ModuleName)
	if err != nil || len(filters) == 0 {
		return fallback
	}

	first := filters[0]
	if first.DataTypeID == nil {
		return fallback
	}

	dataType, err := s.filterRepo.FindDataTypeByID(ctx, *first.DataTypeID)
	if err != nil || dataType == "" {
		return fallback
	}
	return dataType
}

func setStatus(p *domain.UtmLogstashPipeline, status string) {
	p.PipelineStatus = &status
}

func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
