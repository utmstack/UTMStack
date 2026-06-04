package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/incidents/domain"
	"github.com/utmstack/utmstack/backend/pkg/logger"
)

type IncidentMailer interface {
	SendIncidentCreated(ctx context.Context, incident domain.UtmIncident) error
}

type noopMailer struct{}

func (n *noopMailer) SendIncidentCreated(_ context.Context, _ domain.UtmIncident) error {
	logger.Warn("incidents: mail sender not configured — skipping notification")
	return nil
}

func NewNoopMailer() IncidentMailer { return &noopMailer{} }
