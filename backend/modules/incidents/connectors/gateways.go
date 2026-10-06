package connectors

import (
	"context"
	"github.com/utmstack/utmstack/backend/modules/incidents/domain"
	"time"
)

type AlertsGateway interface {
	UpdateAlertStatus(ctx context.Context, alertIDs []string, status domain.IncidentStatus, observation string) error
	MarkAlertAsIncident(ctx context.Context, alertIDs []string, incidentName string, incidentID string, createdAt time.Time, usert string) error

}

type IncidentMailer interface {
	SendIncidentCreated(ctx context.Context, incident domain.Incident) error
}
