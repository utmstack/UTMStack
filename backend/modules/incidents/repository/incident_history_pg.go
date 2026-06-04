package repository

import (
	"context"
	"errors"

	"github.com/utmstack/utmstack/backend/modules/incidents/connectors"
	"github.com/utmstack/utmstack/backend/modules/incidents/domain"
	"github.com/utmstack/utmstack/backend/modules/incidents/dto"
	"gorm.io/gorm"
)

type pgIncidentHistoryRepository struct {
	db *gorm.DB
}

func NewIncidentHistoryRepository(db *gorm.DB) connectors.IncidentHistoryRepository {
	return &pgIncidentHistoryRepository{db: db}
}

func (r *pgIncidentHistoryRepository) Save(ctx context.Context, h *domain.UtmIncidentHistory) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *pgIncidentHistoryRepository) FindByID(ctx context.Context, id int64) (*domain.UtmIncidentHistory, error) {
	var row domain.UtmIncidentHistory
	if err := r.db.WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgIncidentHistoryRepository) FindAll(ctx context.Context, q dto.HistoryListQuery) ([]domain.UtmIncidentHistory, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Size < 1 || q.Size > 200 {
		q.Size = 20
	}

	db := r.db.WithContext(ctx).Model(&domain.UtmIncidentHistory{})

	if q.IncidentID != nil {
		db = db.Where("incident_id = ?", *q.IncidentID)
	}
	if q.ActionType != nil && *q.ActionType != "" {
		db = db.Where("action_type = ?", *q.ActionType)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderBy := "action_created_date DESC"
	if q.Sort != "" {
		orderBy = q.Sort
	}

	var rows []domain.UtmIncidentHistory
	if err := db.Order(orderBy).
		Offset((q.Page - 1) * q.Size).
		Limit(q.Size).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *pgIncidentHistoryRepository) Count(ctx context.Context, q dto.HistoryListQuery) (int64, error) {
	db := r.db.WithContext(ctx).Model(&domain.UtmIncidentHistory{})
	if q.IncidentID != nil {
		db = db.Where("incident_id = ?", *q.IncidentID)
	}
	if q.ActionType != nil && *q.ActionType != "" {
		db = db.Where("action_type = ?", *q.ActionType)
	}
	var total int64
	return total, db.Count(&total).Error
}
