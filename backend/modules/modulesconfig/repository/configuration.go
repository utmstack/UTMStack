package repository

import (
	"context"
	"errors"

	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
	"gorm.io/gorm"
)

type pgConfigRepository struct {
	db *gorm.DB
}

func NewConfigRepository(db *gorm.DB) connectors.ConfigRepository {
	return &pgConfigRepository{db: db}
}

func (r *pgConfigRepository) ListByGroupID(ctx context.Context, groupID int64) ([]domain.UtmModuleGroupConfiguration, error) {
	var items []domain.UtmModuleGroupConfiguration
	err := r.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		Order("conf_key ASC").
		Find(&items).Error
	return items, err
}

func (r *pgConfigRepository) GetByGroupAndKey(ctx context.Context, groupID int64, confKey string) (*domain.UtmModuleGroupConfiguration, error) {
	var c domain.UtmModuleGroupConfiguration
	err := r.db.WithContext(ctx).
		Where("group_id = ? AND conf_key = ?", groupID, confKey).
		First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrConfigNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *pgConfigRepository) SaveAll(ctx context.Context, configs []domain.UtmModuleGroupConfiguration) error {
	if len(configs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range configs {
			if err := tx.Save(&configs[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
