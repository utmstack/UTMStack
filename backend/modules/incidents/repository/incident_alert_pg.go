package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/utmstack/utmstack/backend/modules/incidents/connectors"
	"github.com/utmstack/utmstack/backend/modules/incidents/domain"
	"github.com/utmstack/utmstack/backend/modules/incidents/dto"
	"gorm.io/gorm"
)

type pgIncidentAlertRepository struct {
	db *gorm.DB
}

func NewIncidentAlertRepository(db *gorm.DB) connectors.IncidentAlertRepository {
	return &pgIncidentAlertRepository{db: db}
}

func (r *pgIncidentAlertRepository) Save(ctx context.Context, alert *domain.IncidentAlert) error {
	if alert.TenantID == uuid.Nil {
		alert.TenantID = tenantFromCtx(ctx)
	}
	return r.db.WithContext(ctx).Create(alert).Error
}

func (r *pgIncidentAlertRepository) Update(ctx context.Context, alert *domain.IncidentAlert) error {
	return scopeTenant(ctx, r.db.WithContext(ctx)).Save(alert).Error
}

func (r *pgIncidentAlertRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.IncidentAlert, error) {
	var row domain.IncidentAlert
	if err := scopeTenant(ctx, r.db.WithContext(ctx)).Where("id = ?", id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgIncidentAlertRepository) FindByIncidentID(ctx context.Context, incidentID uuid.UUID) ([]domain.IncidentAlert, error) {
	var rows []domain.IncidentAlert
	if err := scopeTenant(ctx, r.db.WithContext(ctx)).Where("incident_id = ?", incidentID).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *pgIncidentAlertRepository) FindAll(ctx context.Context, q dto.IncidentAlertListQuery) ([]domain.IncidentAlert, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Size < 1 || q.Size > 200 {
		q.Size = 20
	}

	db := scopeTenant(ctx, r.db.WithContext(ctx).Model(&domain.IncidentAlert{}))

	if q.IncidentID != nil {
		db = db.Where("incident_id = ?", *q.IncidentID)
	}
	if q.AlertID != nil && *q.AlertID != "" {
		db = db.Where("alert_id = ?", *q.AlertID)
	}
	if q.AlertStatus != nil && *q.AlertStatus != "" {
		db = db.Where("alert_status = ?", *q.AlertStatus)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []domain.IncidentAlert
	if err := db.Order("alert_severity DESC, alert_name ASC, id ASC").
		Offset((q.Page - 1) * q.Size).
		Limit(q.Size).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *pgIncidentAlertRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return scopeTenant(ctx, r.db.WithContext(ctx).Unscoped()).
		Where("id = ?", id).
		Delete(&domain.IncidentAlert{}).Error
}

func (r *pgIncidentAlertRepository) FindByAlertIDs(ctx context.Context, alertIDs []string) ([]domain.IncidentAlert, error) {
	if len(alertIDs) == 0 {
		return nil, nil
	}
	var rows []domain.IncidentAlert
	if err := scopeTenant(ctx, r.db.WithContext(ctx)).Where("alert_id IN ?", alertIDs).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *pgIncidentAlertRepository) BulkUpdateStatus(ctx context.Context, alertIDs []string, status string) error {
	if len(alertIDs) == 0 {
		return nil
	}
	return scopeTenant(ctx, r.db.WithContext(ctx).Model(&domain.IncidentAlert{})).
		Where("alert_id IN ?", alertIDs).
		Update("alert_status", status).Error
}

func (r *pgIncidentAlertRepository) AlertIDsByIncident(ctx context.Context, incidentID uuid.UUID) ([]string, error) {
	var out []string
	err := scopeTenant(ctx, r.db.WithContext(ctx).Model(&domain.IncidentAlert{})).
		Where("incident_id = ?", incidentID).
		Pluck("alert_id", &out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

// severityRank orders the stored words. Postgres would otherwise compare them
// as text, where "low" sorts above "high" — the inverse of what they mean.
const severityRank = `CASE alert_severity WHEN 'high' THEN 3 WHEN 'medium' THEN 2 WHEN 'low' THEN 1 ELSE 0 END`

func (r *pgIncidentAlertRepository) WorstSeverity(ctx context.Context, incidentID uuid.UUID) (domain.IncidentSeverity, error) {
	var out []string
	err := scopeTenant(ctx, r.db.WithContext(ctx).Model(&domain.IncidentAlert{})).
		Where("incident_id = ?", incidentID).
		Order(severityRank+" DESC").
		Limit(1).
		Pluck("alert_severity", &out).Error
	if err != nil {
		return "", err
	}
	if len(out) == 0 {
		// No alerts left means no severity, which is not a low one.
		return "", nil
	}
	return domain.IncidentSeverity(out[0]), nil
}
