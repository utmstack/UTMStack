package scheduler

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/config"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/internal/association"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/internal/client"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/internal/extractor"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/internal/mapper"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/internal/models"
)

const (
	pollInterval            = 5 * time.Minute
	incidentRetentionPeriod = 48 * time.Hour
	alertRetentionPeriod    = 72 * time.Hour
	cleanupInterval         = 6 * time.Hour
)

type IncidentState struct {
	LastProcessedAt time.Time
	ProcessedAlerts map[string]time.Time
	TotalEntities   int
}

type IngestionScheduler struct {
	cfg                *config.TWConfig
	backendClient      *client.BackendClient
	opensearchClient   *client.OpenSearchClient
	threadwindsClient  *client.ThreadWindsClient
	fieldExtractor     *extractor.FieldExtractor
	entityMapper       *mapper.EntityMapper
	associationBuilder *association.AssociationBuilder
	processedIncidents map[int64]*IncidentState
	mu                 sync.RWMutex
}

func NewIngestionScheduler(
	cfg *config.TWConfig,
	backendClient *client.BackendClient,
	opensearchClient *client.OpenSearchClient,
	threadwindsClient *client.ThreadWindsClient,
) *IngestionScheduler {
	return &IngestionScheduler{
		cfg:                cfg,
		backendClient:      backendClient,
		opensearchClient:   opensearchClient,
		threadwindsClient:  threadwindsClient,
		fieldExtractor:     extractor.NewFieldExtractor(),
		entityMapper:       mapper.NewEntityMapper(),
		associationBuilder: association.NewAssociationBuilder(),
		processedIncidents: make(map[int64]*IncidentState),
	}
}

func (s *IngestionScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	cleanupTicker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	defer cleanupTicker.Stop()

	catcher.Info("ingestion scheduler started", map[string]any{
		"poll_interval":    pollInterval,
		"cleanup_interval": cleanupInterval,
	})

	s.runIngestionCycle(ctx)

	for {
		select {
		case <-ctx.Done():
			catcher.Info("scheduler stopped", nil)
			return
		case <-ticker.C:
			s.runIngestionCycle(ctx)
		case <-cleanupTicker.C:
			s.cleanOldState()
		}
	}
}

func (s *IngestionScheduler) runIngestionCycle(ctx context.Context) {
	catcher.Info("starting ingestion cycle", nil)
	startTime := time.Now()

	if err := s.updateThreadWindsCredentials(ctx); err != nil {
		catcher.Error("failed to update ThreadWinds credentials from database", err, nil)
	}

	cycleTimeout := time.Duration(float64(pollInterval) * 0.9)
	cycleCtx, cancel := context.WithTimeout(ctx, cycleTimeout)
	defer cancel()

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

		entitiesCount, err := s.processIncident(cycleCtx, incident)
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

func (s *IngestionScheduler) processIncident(ctx context.Context, incident *models.Incident) (int, error) {
	s.mu.Lock()
	state, exists := s.processedIncidents[incident.ID]
	if !exists {
		state = &IncidentState{
			ProcessedAlerts: make(map[string]time.Time),
		}
		s.processedIncidents[incident.ID] = state
	}
	s.mu.Unlock()

	incidentAlerts, err := s.backendClient.GetIncidentAlerts(ctx, incident.ID)
	if err != nil {
		return 0, fmt.Errorf("failed to get incident alerts: %w", err)
	}

	if len(incidentAlerts) == 0 {
		return 0, nil
	}

	newAlerts := s.filterNewAlerts(incidentAlerts, state)
	if len(newAlerts) == 0 {
		return 0, nil
	}

	catcher.Info("processing incident with new alerts", map[string]any{
		"incident_id":  incident.ID,
		"new_alerts":   len(newAlerts),
		"total_alerts": len(incidentAlerts),
	})

	s.associationBuilder.ClearRegistry()

	for _, incidentAlert := range newAlerts {
		err := s.processAlertWithAssociations(ctx, incidentAlert, incident)
		if err != nil {
			catcher.Error("failed to process alert", err, map[string]any{
				"alert_id":    incidentAlert.AlertID,
				"incident_id": incident.ID,
			})
			continue
		}

		s.mu.Lock()
		state.ProcessedAlerts[incidentAlert.AlertID] = time.Now()
		s.mu.Unlock()
	}

	allEntities := s.associationBuilder.BuildAssociations()

	if len(allEntities) > 0 {
		if err := s.threadwindsClient.IngestBatch(ctx, allEntities); err != nil {
			return 0, fmt.Errorf("failed to ingest batch: %w", err)
		}
	}

	s.mu.Lock()
	state.LastProcessedAt = time.Now()
	state.TotalEntities += len(allEntities)
	s.mu.Unlock()

	return len(allEntities), nil
}

func (s *IngestionScheduler) processAlertWithAssociations(ctx context.Context, incidentAlert *models.IncidentAlert, incident *models.Incident) error {
	alert, err := s.opensearchClient.GetAlertByID(ctx, incidentAlert.AlertID)
	if err != nil {
		return fmt.Errorf("failed to get alert: %w", err)
	}

	alertFields := s.fieldExtractor.ExtractFromAlert(alert)
	s.mapAndRegisterFieldsToEntities(alertFields, incident, alert)

	return nil
}

func (s *IngestionScheduler) mapAndRegisterFieldsToEntities(
	fields []*models.FlattenedField,
	incident *models.Incident,
	alert *models.Alert,
) {
	for _, field := range fields {
		entityType, matched := s.entityMapper.MapFieldToEntityType(field)
		if !matched {
			continue
		}

		sourceField := ""
		var hostContext *models.Host
		if strings.Contains(field.Path, "source") {
			sourceField = "source"
			hostContext = alert.Source
		} else if strings.Contains(field.Path, "destination") {
			sourceField = "destination"
			hostContext = alert.Destination
		}

		enrichmentCtx := s.buildEnrichmentContext(incident, alert, hostContext)
		entity, entityID, err := s.entityMapper.BuildEntity(entityType, field.Value, enrichmentCtx)
		if err != nil {
			catcher.Error("failed to build entity", err, map[string]any{
				"entity_type": entityType,
				"field_path":  field.Path,
			})
			continue
		}

		assocContext := association.AssociationContext{
			AlertID:     alert.ID,
			IncidentID:  fmt.Sprintf("%d", incident.ID),
			SourceField: sourceField,
		}
		s.associationBuilder.RegisterEntity(entity, entityID, field.Path, assocContext)
	}
}

func (s *IngestionScheduler) filterNewAlerts(incidentAlerts []*models.IncidentAlert, state *IncidentState) []*models.IncidentAlert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	newAlerts := make([]*models.IncidentAlert, 0, len(incidentAlerts))
	for _, alert := range incidentAlerts {
		if _, processed := state.ProcessedAlerts[alert.AlertID]; !processed {
			newAlerts = append(newAlerts, alert)
		}
	}
	return newAlerts
}

func (s *IngestionScheduler) cleanOldState() {
	s.mu.Lock()
	defer s.mu.Unlock()

	incidentCutoff := time.Now().Add(-incidentRetentionPeriod)
	alertCutoff := time.Now().Add(-alertRetentionPeriod)

	cleanedIncidents := 0
	cleanedAlerts := 0

	for incidentID, state := range s.processedIncidents {
		if state.LastProcessedAt.Before(incidentCutoff) {
			delete(s.processedIncidents, incidentID)
			cleanedIncidents++
			continue
		}

		for alertID, processedAt := range state.ProcessedAlerts {
			if processedAt.Before(alertCutoff) {
				delete(state.ProcessedAlerts, alertID)
				cleanedAlerts++
			}
		}
	}

	if cleanedIncidents > 0 || cleanedAlerts > 0 {
		catcher.Info("state cleanup completed", map[string]any{
			"active_incidents":  len(s.processedIncidents),
			"cleaned_incidents": cleanedIncidents,
			"cleaned_alerts":    cleanedAlerts,
		})
	}
}

func (s *IngestionScheduler) buildEnrichmentContext(incident *models.Incident, alert *models.Alert, host *models.Host) mapper.EntityEnrichmentContext {
	ctx := mapper.EntityEnrichmentContext{
		IncidentID: fmt.Sprintf("%d", incident.ID),
		Severity:   incident.Severity,
		DataType:   alert.DataType,
	}

	if host != nil {
		ctx.Country = host.Country
		ctx.City = host.City
		ctx.ASO = host.ASO

		if len(host.Coordinates) == 2 {
			lat := host.Coordinates[0]
			lon := host.Coordinates[1]
			if lat != 0.0 || lon != 0.0 {
				ctx.Latitude = &lat
				ctx.Longitude = &lon
			}
		}

		if host.AccuracyRadius > 0 {
			radius := float64(host.AccuracyRadius)
			ctx.AccuracyRadius = &radius
		}
	}

	return ctx
}

func (s *IngestionScheduler) updateThreadWindsCredentials(ctx context.Context) error {
	config, err := s.backendClient.GetThreadWindsConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get ThreadWinds credentials from backend: %w", err)
	}

	if config.APIKey == "" || config.APISecret == "" {
		return fmt.Errorf("ThreadWinds credentials are empty in backend configuration")
	}

	s.threadwindsClient.UpdateCredentials(config.APIKey, config.APISecret)

	return nil
}
