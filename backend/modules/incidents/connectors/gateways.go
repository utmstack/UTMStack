package connectors

import (
	"context"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/incidents/domain"
	"github.com/utmstack/utmstack/backend/modules/incidents/dto"
)

// Outbound ports — the abstractions through which incidents talks to the other
// modules it depends on (alerts, iam, mail). Concrete adapters that bridge these
// to the real modules live in the composition layer (package incidents), so this
// ports package stays free of cross-module concrete imports.

// AlertsGateway syncs alert status changes back to the alerts module (OpenSearch).
type AlertsGateway interface {
	UpdateAlertStatus(ctx context.Context, alertIDs []string, status int, observation string) error
}

type noopAlertsGateway struct{}

func (n *noopAlertsGateway) UpdateAlertStatus(_ context.Context, _ []string, _ int, _ string) error {
	catcher.Warn("incidents: AlertsGateway not configured — OpenSearch alert status sync skipped", nil)
	return nil
}

func NewNoopAlertsGateway() AlertsGateway { return &noopAlertsGateway{} }

// IAMGateway resolves assigned users from the iam module.
type IAMGateway interface {
	FindUsersByIDs(ctx context.Context, ids []int64) ([]dto.UserAssignedDTO, error)
}

type noopIAMGateway struct{}

func (n *noopIAMGateway) FindUsersByIDs(_ context.Context, ids []int64) ([]dto.UserAssignedDTO, error) {
	result := make([]dto.UserAssignedDTO, 0, len(ids))
	for _, id := range ids {
		result = append(result, dto.UserAssignedDTO{ID: id, Login: ""})
	}
	return result, nil
}

func NewNoopIAMGateway() IAMGateway { return &noopIAMGateway{} }

// IncidentMailer sends incident notifications.
type IncidentMailer interface {
	SendIncidentCreated(ctx context.Context, incident domain.UtmIncident) error
}

type noopMailer struct{}

func (n *noopMailer) SendIncidentCreated(_ context.Context, _ domain.UtmIncident) error {
	catcher.Warn("incidents: mail sender not configured — skipping notification", nil)
	return nil
}

func NewNoopMailer() IncidentMailer { return &noopMailer{} }
