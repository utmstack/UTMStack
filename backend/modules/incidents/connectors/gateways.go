package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/incidents/domain"
)

type AlertsGateway interface {
	UpdateAlertStatus(ctx context.Context, alertIDs []string, status domain.IncidentStatus, observation string) error
}

type IncidentMailer interface {
	SendIncidentCreated(ctx context.Context, incident domain.Incident) error
}
