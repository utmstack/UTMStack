package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/incidents/dto"
)

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
