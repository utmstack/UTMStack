package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
	"gorm.io/gorm"
)

var _ connectors.ActionRepository = (*actionRepository)(nil)

type actionRepository struct {
	db *gorm.DB
}

func NewActionRepository(db *gorm.DB) *actionRepository {
	return &actionRepository{db: db}
}

func (r *actionRepository) Save(ctx context.Context, action *domain.UtmIncidentAction) error {
	if action.TenantID == "" {
		action.TenantID = tenantFromCtx(ctx)
	}
	if err := r.db.Save(action).Error; err != nil {
		return fmt.Errorf("actionRepository.Save: %w", err)
	}
	return nil
}

func (r *actionRepository) FindByID(ctx context.Context, id int64) (*domain.UtmIncidentAction, error) {
	var a domain.UtmIncidentAction
	err := scopeTenant(ctx, r.db).First(&a, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrIncidentRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("actionRepository.FindByID: %w", err)
	}
	return &a, nil
}

func (r *actionRepository) FindAll(ctx context.Context, f dto.ActionFilter) ([]domain.UtmIncidentAction, int64, error) {
	q := scopeTenant(ctx, r.db.Model(&domain.UtmIncidentAction{}))
	if f.ActionCommand != nil {
		q = q.Where("action_command = ?", *f.ActionCommand)
	}
	if f.ActionType != nil {
		q = q.Where("action_type = ?", *f.ActionType)
	}
	if f.ActionEditable != nil {
		q = q.Where("action_editable = ?", *f.ActionEditable)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("actionRepository.FindAll count: %w", err)
	}

	var items []domain.UtmIncidentAction
	if err := q.Offset(f.Offset()).Limit(f.Limit()).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("actionRepository.FindAll: %w", err)
	}
	return items, total, nil
}

func (r *actionRepository) Delete(ctx context.Context, id int64) error {
	if err := scopeTenant(ctx, r.db).Delete(&domain.UtmIncidentAction{}, id).Error; err != nil {
		return fmt.Errorf("actionRepository.Delete: %w", err)
	}
	return nil
}
