package connectors

import (
	"context"

	alerts_connectors "github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	alerts_dto "github.com/utmstack/utmstack/backend/modules/alerts/dto"
)

type alertsGatewayAdapter struct {
	uc alerts_connectors.AlertUsecase
}

func NewAlertsGatewayFromUsecase(uc alerts_connectors.AlertUsecase) AlertsGateway {
	return &alertsGatewayAdapter{uc: uc}
}

func (a *alertsGatewayAdapter) UpdateAlertStatus(ctx context.Context, alertIDs []string, status int, observation string) error {
	// TODO(incidents): thread the acting user through AlertsGateway so the alert
	// history attributes incident-driven status changes to the real user instead
	// of "system". Empty login → alerts usecase resolves it to "system".
	return a.uc.UpdateStatus(ctx, "", alerts_dto.UpdateAlertStatusRequest{
		AlertIDs:          alertIDs,
		Status:            status,
		StatusObservation: observation,
	})
}
