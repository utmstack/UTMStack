package scheduler

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/feeds/config"
	"github.com/utmstack/UTMStack/plugins/feeds/internal/client"
	"github.com/utmstack/UTMStack/plugins/feeds/internal/service"
)

const (
	pollInterval = 5 * time.Minute
)

type IngestionScheduler struct {
	cfg               *config.TWConfig
	backendClient     *client.BackendClient
	threadwindsClient *client.ThreadWindsClient
	incidentProcessor *service.IncidentProcessor
}

func NewIngestionScheduler(
	cfg *config.TWConfig,
	deps *client.ClientDependencies,
	incidentProcessor *service.IncidentProcessor,
) *IngestionScheduler {
	return &IngestionScheduler{
		cfg:               cfg,
		backendClient:     deps.Backend,
		threadwindsClient: deps.ThreadWinds,
		incidentProcessor: incidentProcessor,
	}
}

func (s *IngestionScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	catcher.Info("ingestion scheduler started", map[string]any{
		"poll_interval": pollInterval,
	})

	s.runIngestionCycle(ctx)

	for {
		select {
		case <-ctx.Done():
			catcher.Info("scheduler received shutdown signal, stopping gracefully", nil)
			return
		case <-ticker.C:
			s.runIngestionCycle(ctx)
		}
	}
}

func (s *IngestionScheduler) runIngestionCycle(ctx context.Context) {
	startTime := time.Now()

	cycleTimeout := time.Duration(float64(pollInterval) * 0.9)
	cycleCtx, cancel := context.WithTimeout(ctx, cycleTimeout)
	defer cancel()

	twConfig, err := s.backendClient.GetThreadWindsConfig(cycleCtx)
	if err != nil {
		catcher.Error("failed to get ThreadWinds configuration", err, nil)
		return
	}

	if twConfig.Enabled != "true" {
		catcher.Info("ThreadWinds is disabled, skipping ingestion cycle", nil)
		return
	}

	if twConfig.APIKey != "" && twConfig.APISecret != "" {
		s.threadwindsClient.UpdateCredentials(twConfig.APIKey, twConfig.APISecret)
	}

	incidents, err := s.backendClient.GetRecentIncidents(cycleCtx)
	if err != nil {
		catcher.Error("failed to fetch incidents", err, nil)
		return
	}

	if len(incidents) == 0 {
		catcher.Info("no recent incidents to process", nil)
		return
	}

	totalEntities := 0
	for i, incident := range incidents {
		select {
		case <-cycleCtx.Done():
			catcher.Info("cycle timeout or cancellation, stopping", map[string]any{
				"processed_incidents": i,
				"total_incidents":     len(incidents),
				"reason":              cycleCtx.Err().Error(),
			})
			return
		default:
		}

		entitiesCount, err := s.incidentProcessor.ProcessIncident(cycleCtx, incident)
		if err != nil {
			catcher.Error("failed to process incident", err, map[string]any{
				"incident_id":   incident.ID,
				"incident_name": incident.Name,
			})
			continue
		}
		totalEntities += entitiesCount
	}

	duration := time.Since(startTime)
	catcher.Info("ingestion cycle completed", map[string]any{
		"duration_seconds":    duration.Seconds(),
		"incidents_processed": len(incidents),
		"total_entities":      totalEntities,
	})
}
