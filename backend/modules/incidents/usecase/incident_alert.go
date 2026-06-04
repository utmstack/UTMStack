package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/incidents/connectors"
	"github.com/utmstack/utmstack/backend/modules/incidents/domain"
	"github.com/utmstack/utmstack/backend/modules/incidents/dto"
)

type incidentAlertUsecase struct {
	alertRepo   connectors.IncidentAlertRepository
	historyRepo connectors.IncidentHistoryRepository
}

func NewIncidentAlertUsecase(
	alertRepo connectors.IncidentAlertRepository,
	historyRepo connectors.IncidentHistoryRepository,
) connectors.IncidentAlertUsecase {
	return &incidentAlertUsecase{
		alertRepo:   alertRepo,
		historyRepo: historyRepo,
	}
}

func (u *incidentAlertUsecase) Create(ctx context.Context, req dto.IncidentAlertRequest) (*domain.UtmIncidentAlert, error) {
	row := &domain.UtmIncidentAlert{
		IncidentID:    req.IncidentID,
		AlertID:       req.AlertID,
		AlertName:     req.AlertName,
		AlertSeverity: req.AlertSeverity,
		AlertStatus:   req.AlertStatus,
	}
	if err := u.alertRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (u *incidentAlertUsecase) UpdateStatus(ctx context.Context, userLogin string, req dto.UpdateAlertStatusRequest) error {
	if err := u.alertRepo.BulkUpdateStatus(ctx, req.AlertIds, req.AlertStatus); err != nil {
		return err
	}

	currentUser := resolveUser(userLogin)

	statusLabel := domain.AlertStatusLabel(req.AlertStatus)
	detail := fmt.Sprintf("Alerts changed to %s", statusLabel)

	now := time.Now().UTC()
	by := currentUser
	h := &domain.UtmIncidentHistory{
		IncidentID:        req.IncidentID,
		Action:            domain.ActionAlertStatusChanged.Label,
		ActionType:        domain.ActionAlertStatusChanged.Type,
		ActionDetail:      &detail,
		ActionCreatedDate: now,
		ActionCreatedBy:   &by,
	}
	if saveErr := u.historyRepo.Save(ctx, h); saveErr != nil {
		catcher.Warn("incidents: failed to write alert-status history", map[string]any{"error": saveErr.Error()})
	}

	return nil
}

func (u *incidentAlertUsecase) Update(ctx context.Context, req dto.UpdateIncidentAlertRequest) (*domain.UtmIncidentAlert, error) {
	row := &domain.UtmIncidentAlert{
		ID:            req.ID,
		IncidentID:    req.IncidentID,
		AlertID:       req.AlertID,
		AlertName:     req.AlertName,
		AlertSeverity: req.AlertSeverity,
		AlertStatus:   req.AlertStatus,
	}
	if err := u.alertRepo.Update(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (u *incidentAlertUsecase) List(ctx context.Context, query dto.IncidentAlertListQuery) ([]domain.UtmIncidentAlert, int64, error) {
	return u.alertRepo.FindAll(ctx, query)
}

func (u *incidentAlertUsecase) Delete(ctx context.Context, id int64) error {
	return u.alertRepo.Delete(ctx, id)
}
