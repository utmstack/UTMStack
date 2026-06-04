package repository

import (
	"context"
	"errors"

	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
	"gorm.io/gorm"
)

type pgGroupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) connectors.GroupRepository {
	return &pgGroupRepository{db: db}
}

func (r *pgGroupRepository) GetByID(ctx context.Context, id int64) (*domain.UtmModuleGroup, error) {
	var g domain.UtmModuleGroup
	err := r.db.WithContext(ctx).
		Preload("ModuleGroupConfigurations").
		First(&g, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrGroupNotFound
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *pgGroupRepository) ListByModuleID(ctx context.Context, moduleID int64) ([]domain.UtmModuleGroup, error) {
	var groups []domain.UtmModuleGroup
	err := r.db.WithContext(ctx).
		Preload("ModuleGroupConfigurations").
		Where("module_id = ?", moduleID).
		Order("group_name ASC").
		Find(&groups).Error
	return groups, err
}

func (r *pgGroupRepository) Save(ctx context.Context, group *domain.UtmModuleGroup) error {
	return r.db.WithContext(ctx).Save(group).Error
}

func (r *pgGroupRepository) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&domain.UtmModuleGroup{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrGroupNotFound
	}
	return nil
}

func (r *pgGroupRepository) DeleteAllByModuleID(ctx context.Context, moduleID int64) error {
	return r.db.WithContext(ctx).
		Where("module_id = ?", moduleID).
		Delete(&domain.UtmModuleGroup{}).Error
}
