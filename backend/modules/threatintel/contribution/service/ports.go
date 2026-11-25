package service

import (
	"context"

	"github.com/threatwinds/go-sdk/entities"

	"github.com/utmstack/utmstack/backend/modules/threatintel/contribution/models"
)

// IncidentSource gives the incidents to contribute and the alerts of each.
type IncidentSource interface {
	// RecentIncidents returns the incidents that are open or under review.
	RecentIncidents(ctx context.Context) ([]*models.Incident, error)
	// IncidentAlerts returns the alerts linked to an incident.
	IncidentAlerts(ctx context.Context, incidentID string) ([]*models.IncidentAlert, error)
}

// AlertSource reads an alert, with its events, by id.
type AlertSource interface {
	AlertByID(ctx context.Context, alertID string) (*models.Alert, error)
}

// Ingestor sends entities to the intelligence platform.
type Ingestor interface {
	IngestBatch(ctx context.Context, batch []*entities.Entity) models.IngestOutcome
}
