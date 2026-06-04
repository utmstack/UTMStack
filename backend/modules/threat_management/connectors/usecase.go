package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/threat_management/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type AdversaryUsecase interface {
	FetchAdversaryAlerts(ctx context.Context, filters []common_models.FilterType) ([]dto.AdversaryResponse, error)
}
