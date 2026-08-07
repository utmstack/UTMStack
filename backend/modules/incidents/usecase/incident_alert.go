package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/incidents/connectors"
	"github.com/utmstack/utmstack/backend/modules/incidents/domain"
	"github.com/utmstack/utmstack/backend/modules/incidents/dto"
)

type incidentAlertUsecase struct {
	alertRepo    connectors.IncidentAlertRepository
	historyRepo  connectors.IncidentHistoryRepository
	incidentRepo connectors.IncidentRepository
}

func NewIncidentAlertUsecase(
	alertRepo connectors.IncidentAlertRepository,
	historyRepo connectors.IncidentHistoryRepository,
	incidentRepo connectors.IncidentRepository,
) connectors.IncidentAlertUsecase {
	return &incidentAlertUsecase{
		alertRepo:    alertRepo,
		historyRepo:  historyRepo,
		incidentRepo: incidentRepo,
	}
}

func (u *incidentAlertUsecase) Create(ctx context.Context, req dto.IncidentAlertRequest) (*domain.IncidentAlert, error) {
	severity := domain.IncidentSeverity(req.AlertSeverity)
	if !severity.Valid() {
		return nil, fmt.Errorf("alert %s: unknown severity %q", req.AlertID, req.AlertSeverity)
	}

	row := &domain.IncidentAlert{
		IncidentID:    req.IncidentID,
		AlertID:       req.AlertID,
		AlertName:     req.AlertName,
		AlertSeverity: severity,
	}
	if req.AlertStatus != nil {
		row.AlertStatus = *req.AlertStatus
	}
	if err := u.alertRepo.Save(ctx, row); err != nil {
		return nil, err
	}

	u.recomputeSeverity(ctx, req.IncidentID)
	return row, nil
}

func (u *incidentAlertUsecase) UpdateStatus(ctx context.Context, userEmail string, req dto.UpdateAlertStatusRequest) error {
	if err := u.alertRepo.BulkUpdateStatus(ctx, req.AlertIds, req.AlertStatus); err != nil {
		return err
	}

	by := resolveUser(userEmail)
	h := &domain.IncidentHistory{
		IncidentID:        req.IncidentID,
		Action:            domain.ActionAlertStatusChanged,
		ActionCreatedDate: time.Now().UTC(),
		ActionCreatedBy:   &by,
	}
	if saveErr := u.historyRepo.Save(ctx, h); saveErr != nil {
		catcher.Warn("incidents: failed to write alert-status history", map[string]any{"error": saveErr.Error()})
	}

	return nil
}

func (u *incidentAlertUsecase) Update(ctx context.Context, req dto.UpdateIncidentAlertRequest) (*domain.IncidentAlert, error) {
	severity := domain.IncidentSeverity(req.AlertSeverity)
	if !severity.Valid() {
		return nil, fmt.Errorf("alert %s: unknown severity %q", req.AlertID, req.AlertSeverity)
	}

	row := &domain.IncidentAlert{
		ID:            req.ID,
		IncidentID:    req.IncidentID,
		AlertID:       req.AlertID,
		AlertName:     req.AlertName,
		AlertSeverity: severity,
	}
	if req.AlertStatus != nil {
		row.AlertStatus = *req.AlertStatus
	}
	if err := u.alertRepo.Update(ctx, row); err != nil {
		return nil, err
	}
	u.recomputeSeverity(ctx, req.IncidentID)
	return row, nil
}

func (u *incidentAlertUsecase) List(ctx context.Context, query dto.IncidentAlertListQuery) ([]domain.IncidentAlert, int64, error) {
	return u.alertRepo.FindAll(ctx, query)
}

func (u *incidentAlertUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	row, err := u.alertRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := u.alertRepo.Delete(ctx, id); err != nil {
		return err
	}
	if row != nil {
		u.recomputeSeverity(ctx, row.IncidentID)
	}
	return nil
}

// The database decides the worst one and returns a single row. Folding every
// alert of the incident in Go ran on each link, edit and unlink, and the cost
// grew with the incident.
func (u *incidentAlertUsecase) recomputeSeverity(ctx context.Context, incidentID uuid.UUID) {
	if u.incidentRepo == nil {
		return
	}
	worst, err := u.alertRepo.WorstSeverity(ctx, incidentID)
	if err != nil {
		catcher.Warn("incidents: failed to read the worst severity", map[string]any{"error": err.Error()})
		return
	}
	incident, err := u.incidentRepo.FindByID(ctx, incidentID)
	if err != nil || incident == nil {
		return
	}
	if worst == incident.Severity {
		return
	}
	incident.Severity = worst
	if err := u.incidentRepo.Update(ctx, incident); err != nil {
		catcher.Warn("incidents: failed to recompute severity", map[string]any{"error": err.Error()})
	}
}
