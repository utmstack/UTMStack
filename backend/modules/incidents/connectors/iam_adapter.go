package connectors

import (
	"context"

	iam_connectors "github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/incidents/dto"
)

type iamGatewayAdapter struct {
	userRepo iam_connectors.UserRepository
}

func NewIAMGatewayFromRepo(userRepo iam_connectors.UserRepository) IAMGateway {
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
