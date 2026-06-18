package repository

import (
	"context"
	"errors"

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

func (r *pgIncidentAlertRepository) Save(ctx context.Context, alert *domain.UtmIncidentAlert) error {
	return r.db.WithContext(ctx).Create(alert).Error
}

func (r *pgIncidentAlertRepository) Update(ctx context.Context, alert *domain.UtmIncidentAlert) error {
	return r.db.WithContext(ctx).Save(alert).Error
}

func (r *pgIncidentAlertRepository) FindByID(ctx context.Context, id int64) (*domain.UtmIncidentAlert, error) {
	var row domain.UtmIncidentAlert
	if err := r.db.WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgIncidentAlertRepository) FindByIncidentID(ctx context.Context, incidentID int64) ([]domain.UtmIncidentAlert, error) {
	var rows []domain.UtmIncidentAlert
	if err := r.db.WithContext(ctx).Where("incident_id = ?", incidentID).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *pgIncidentAlertRepository) FindAll(ctx context.Context, q dto.IncidentAlertListQuery) ([]domain.UtmIncidentAlert, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Size < 1 || q.Size > 200 {
		q.Size = 20
	}

	db := r.db.WithContext(ctx).Model(&domain.UtmIncidentAlert{})

	if q.IncidentID != nil {
		db = db.Where("incident_id = ?", *q.IncidentID)
	}
	if q.AlertID != nil && *q.AlertID != "" {
		db = db.Where("alert_id = ?", *q.AlertID)
	}
	if q.AlertStatus != nil {
		db = db.Where("alert_status = ?", *q.AlertStatus)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []domain.UtmIncidentAlert
	if err := db.Order("id ASC").
		Offset((q.Page - 1) * q.Size).
		Limit(q.Size).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *pgIncidentAlertRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&domain.UtmIncidentAlert{}, id).Error
}

func (r *pgIncidentAlertRepository) FindByAlertIDs(ctx context.Context, alertIDs []string) ([]domain.UtmIncidentAlert, error) {
	if len(alertIDs) == 0 {
		return nil, nil
	}
	var rows []domain.UtmIncidentAlert
	if err := r.db.WithContext(ctx).Where("alert_id IN ?", alertIDs).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *pgIncidentAlertRepository) ExistsByAlertID(ctx context.Context, alertID string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.UtmIncidentAlert{}).
		Where("alert_id = ?", alertID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *pgIncidentAlertRepository) BulkUpdateStatus(ctx context.Context, alertIDs []string, status int) error {
	if len(alertIDs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Exec("UPDATE utm_incident_alert SET alert_status = ? WHERE alert_id IN ?", status, alertIDs).
		Error
}
