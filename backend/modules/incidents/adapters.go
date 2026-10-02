package incidents

import (
	"context"
	"time"
	alerts_connectors "github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	alerts_domain "github.com/utmstack/utmstack/backend/modules/alerts/domain"
	alerts_dto "github.com/utmstack/utmstack/backend/modules/alerts/dto"
	"github.com/utmstack/utmstack/backend/modules/incidents/connectors"
	"github.com/utmstack/utmstack/backend/modules/incidents/domain"
)

type alertsGatewayAdapter struct {
	uc alerts_connectors.AlertUsecase
}

func NewAlertsGatewayFromUsecase(uc alerts_connectors.AlertUsecase) connectors.AlertsGateway {
	return &alertsGatewayAdapter{uc: uc}
}

func (a *alertsGatewayAdapter) UpdateAlertStatus(ctx context.Context, alertIDs []string, status domain.IncidentStatus, observation string) error {
	// TODO(incidents): thread the acting user through AlertsGateway so the alert
	// history attributes incident-driven status changes to the real user instead
	// of "system". An empty email → alerts usecase resolves it to "system".
	return a.uc.UpdateStatus(ctx, "", alerts_dto.UpdateAlertStatusRequest{
		AlertIDs:          alertIDs,
		Status:            alerts_domain.AlertStatus(status),
		StatusObservation: observation,
	})
}



func(a *alertsGatewayAdapter) MarkAlertAsIncident(ctx context.Context, alertIDs []string, incidentName string, incidentID string, createdAt time.Time, userEmail string) error{
	req:= alerts_dto.ConvertToIncidentRequest{
		AlertIDs      :alertIDs,
		IncidentName  :incidentName,
		IncidentID    :incidentID,
		IncidentSource:"incidents",
	}
	return a.uc.ConvertToIncident(ctx, userEmail, req)
}
