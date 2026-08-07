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

type pgIncidentHistoryRepository struct {
	db *gorm.DB
}

func NewIncidentHistoryRepository(db *gorm.DB) connectors.IncidentHistoryRepository {
	return &pgIncidentHistoryRepository{db: db}
}

func (r *pgIncidentHistoryRepository) Save(ctx context.Context, h *domain.IncidentHistory) error {
	if h.TenantID == uuid.Nil {
		h.TenantID = tenantFromCtx(ctx)
	}
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *pgIncidentHistoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.IncidentHistory, error) {
	var row domain.IncidentHistory
	if err := scopeTenant(ctx, r.db.WithContext(ctx)).Where("id = ?", id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func filterHistory(db *gorm.DB, q dto.HistoryListQuery) *gorm.DB {
	if q.IncidentID != nil {
		db = db.Where("incident_id = ?", *q.IncidentID)
	}
	if q.Action != nil && *q.Action != "" {
		db = db.Where("action = ?", *q.Action)
	}
	return db
}

func (r *pgIncidentHistoryRepository) FindAll(ctx context.Context, q dto.HistoryListQuery) ([]domain.IncidentHistory, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Size < 1 || q.Size > 200 {
		q.Size = 20
	}

	db := filterHistory(scopeTenant(ctx, r.db.WithContext(ctx).Model(&domain.IncidentHistory{})), q)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []domain.IncidentHistory
	if err := db.Order(orderBy(q.Sort, historySortable, "action_created_date DESC, id ASC")).
		Offset((q.Page - 1) * q.Size).
		Limit(q.Size).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *pgIncidentHistoryRepository) Count(ctx context.Context, q dto.HistoryListQuery) (int64, error) {
	db := filterHistory(scopeTenant(ctx, r.db.WithContext(ctx).Model(&domain.IncidentHistory{})), q)
	var total int64
	return total, db.Count(&total).Error
}
