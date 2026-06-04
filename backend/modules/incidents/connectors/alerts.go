package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/pkg/logger"
)

type AlertsGateway interface {
	UpdateAlertStatus(ctx context.Context, alertIDs []string, status int, observation string) error
}

type noopAlertsGateway struct{}

func (n *noopAlertsGateway) UpdateAlertStatus(_ context.Context, _ []string, _ int, _ string) error {
	logger.Warn("incidents: AlertsGateway not configured — OpenSearch alert status sync skipped")
	return nil
}

func NewNoopAlertsGateway() AlertsGateway { return &noopAlertsGateway{} }
