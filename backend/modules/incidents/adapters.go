package incidents

import (
	"context"

	alerts_connectors "github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	alerts_dto "github.com/utmstack/utmstack/backend/modules/alerts/dto"
	iam_connectors "github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/incidents/connectors"
	"github.com/utmstack/utmstack/backend/modules/incidents/dto"
)

// Adapters bridge the incidents outbound ports (connectors.AlertsGateway /
// IAMGateway) to the concrete sibling modules. They live in the composition
// layer so the connectors (ports) package stays free of cross-module imports.

type alertsGatewayAdapter struct {
	uc alerts_connectors.AlertUsecase
}

func NewAlertsGatewayFromUsecase(uc alerts_connectors.AlertUsecase) connectors.AlertsGateway {
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

type iamGatewayAdapter struct {
	userRepo iam_connectors.UserRepository
}

func NewIAMGatewayFromRepo(userRepo iam_connectors.UserRepository) connectors.IAMGateway {
	return &iamGatewayAdapter{userRepo: userRepo}
}

func (a *iamGatewayAdapter) FindUsersByIDs(ctx context.Context, ids []int64) ([]dto.UserAssignedDTO, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	users, _, err := a.userRepo.List(ctx, iam_connectors.ListUsersFilter{
		IDs:      ids,
		PageSize: len(ids),
	})
	if err != nil {
		return nil, err
	}
	result := make([]dto.UserAssignedDTO, 0, len(users))
	for _, u := range users {
		result = append(result, dto.UserAssignedDTO{
			ID:    int64(u.ID),
			Login: u.Login,
		})
	}
	return result, nil
}
