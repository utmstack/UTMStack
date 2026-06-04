package usecase

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/incidents/connectors"
	"github.com/utmstack/utmstack/backend/modules/incidents/domain"
	incidenterrors "github.com/utmstack/utmstack/backend/modules/incidents/errors"
	"github.com/utmstack/utmstack/backend/modules/incidents/dto"
)

type incidentHistoryUsecase struct {
	historyRepo connectors.IncidentHistoryRepository
}

func NewIncidentHistoryUsecase(historyRepo connectors.IncidentHistoryRepository) connectors.IncidentHistoryUsecase {
	return &incidentHistoryUsecase{historyRepo: historyRepo}
}

func (u *incidentHistoryUsecase) List(ctx context.Context, query dto.HistoryListQuery) ([]domain.UtmIncidentHistory, int64, error) {
	return u.historyRepo.FindAll(ctx, query)
}

func (u *incidentHistoryUsecase) Count(ctx context.Context, query dto.HistoryListQuery) (int64, error) {
	return u.historyRepo.Count(ctx, query)
}

func (u *incidentHistoryUsecase) GetByID(ctx context.Context, id int64) (*domain.UtmIncidentHistory, error) {
	h, err := u.historyRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if h == nil {
		return nil, incidenterrors.ErrNotFound
	}
	return h, nil
}
