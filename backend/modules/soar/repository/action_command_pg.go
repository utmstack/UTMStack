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

var _ connectors.ActionCommandRepository = (*actionCommandRepository)(nil)

type actionCommandRepository struct {
	db *gorm.DB
}

func NewActionCommandRepository(db *gorm.DB) *actionCommandRepository {
	return &actionCommandRepository{db: db}
}

func (r *actionCommandRepository) Save(ctx context.Context, cmd *domain.UtmIncidentActionCommand) error {
	if cmd.TenantID == "" {
		cmd.TenantID = tenantFromCtx(ctx)
	}
	if err := r.db.Save(cmd).Error; err != nil {
		return fmt.Errorf("actionCommandRepository.Save: %w", err)
	}
	return nil
}

func (r *actionCommandRepository) FindByID(ctx context.Context, id int64) (*domain.UtmIncidentActionCommand, error) {
	var c domain.UtmIncidentActionCommand
	err := scopeTenant(ctx, r.db).First(&c, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrIncidentRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("actionCommandRepository.FindByID: %w", err)
	}
	return &c, nil
}

func (r *actionCommandRepository) FindAll(ctx context.Context, f dto.ActionCommandFilter) ([]domain.UtmIncidentActionCommand, int64, error) {
	q := scopeTenant(ctx, r.db.Model(&domain.UtmIncidentActionCommand{}))
	if f.ActionID != nil {
		q = q.Where("action_id = ?", *f.ActionID)
	}
	if f.OsPlatform != nil {
		q = q.Where("os_platform = ?", *f.OsPlatform)
	}
	if f.Command != nil {
		q = q.Where("command ILIKE ?", "%"+*f.Command+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("actionCommandRepository.FindAll count: %w", err)
	}

	var items []domain.UtmIncidentActionCommand
	if err := q.Offset(f.Offset()).Limit(f.Limit()).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("actionCommandRepository.FindAll: %w", err)
	}
	return items, total, nil
}

func (r *actionCommandRepository) Delete(ctx context.Context, id int64) error {
	if err := scopeTenant(ctx, r.db).Delete(&domain.UtmIncidentActionCommand{}, id).Error; err != nil {
		return fmt.Errorf("actionCommandRepository.Delete: %w", err)
	}
	return nil
}
