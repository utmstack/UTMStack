package repository

import (
	"context"
	"errors"

	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
	"gorm.io/gorm"
)

type pgModuleRepository struct {
	db *gorm.DB
}

func NewModuleRepository(db *gorm.DB) connectors.ModuleRepository {
	return &pgModuleRepository{db: db}
}

func (r *pgModuleRepository) GetByID(ctx context.Context, id int64) (*domain.UtmModule, error) {
	var m domain.UtmModule
	err := r.db.WithContext(ctx).
		Preload("ModuleGroups.ModuleGroupConfigurations").
		First(&m, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrModuleNotFound
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *pgModuleRepository) GetByServerAndName(ctx context.Context, serverID int64, moduleName string) (*domain.UtmModule, error) {
	var m domain.UtmModule
	err := r.db.WithContext(ctx).
		Preload("ModuleGroups.ModuleGroupConfigurations").
		Where("server_id = ? AND module_name = ?", serverID, moduleName).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrModuleNotFound
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *pgModuleRepository) List(ctx context.Context, filter connectors.ModuleListFilter) ([]domain.UtmModule, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.UtmModule{})

	if filter.ServerID != nil {
		q = q.Where("server_id = ?", *filter.ServerID)
	}
	if filter.ModuleCategory != nil {
		q = q.Where("module_category = ?", *filter.ModuleCategory)
	}
	if filter.ModuleActive != nil {
		q = q.Where("module_active = ?", *filter.ModuleActive)
	}
	if filter.NameContains != nil && *filter.NameContains != "" {
		q = q.Where("module_name ILIKE ?", "%"+*filter.NameContains+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	size := filter.Size
	if size <= 0 {
		size = 20
	}

	var items []domain.UtmModule
	err := q.
		Preload("ModuleGroups.ModuleGroupConfigurations").
		Order("module_name ASC").
		Limit(size).
		Offset((page - 1) * size).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *pgModuleRepository) Save(ctx context.Context, module *domain.UtmModule) error {
	return r.db.WithContext(ctx).Save(module).Error
}

func (r *pgModuleRepository) Categories(ctx context.Context, serverID *int64) ([]string, error) {
	q := r.db.WithContext(ctx).
		Model(&domain.UtmModule{}).
		Distinct("module_category").
		Where("module_category <> ''")
	if serverID != nil {
		q = q.Where("server_id = ?", *serverID)
	}
	var categories []string
	if err := q.Order("module_category ASC").Pluck("module_category", &categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *pgModuleRepository) CountActiveByName(ctx context.Context, moduleName string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.UtmModule{}).
		Where("module_name = ? AND module_active = ?", moduleName, true).
		Count(&count).Error
	return count, err
}
