package repository

import (
	"context"
	"errors"

	"github.com/utmstack/utmstack/backend/modules/integrations/connectors"
	"github.com/utmstack/utmstack/backend/modules/integrations/domain"
	"gorm.io/gorm"
)

type pgModuleRepository struct {
	db *gorm.DB
}

// NewModuleRepository builds a GORM-backed ModuleRepository over utm_module.
func NewModuleRepository(db *gorm.DB) connectors.ModuleRepository {
	return &pgModuleRepository{db: db}
}

var _ connectors.ModuleRepository = (*pgModuleRepository)(nil)

func (r *pgModuleRepository) GetByID(ctx context.Context, id int64) (*domain.UtmModule, error) {
	var m domain.UtmModule
	err := r.db.WithContext(ctx).First(&m, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrModuleNotFound
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *pgModuleRepository) GetByName(ctx context.Context, moduleName string) (*domain.UtmModule, error) {
	var m domain.UtmModule
	err := r.db.WithContext(ctx).
		Where("module_name = ?", moduleName).
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

func (r *pgModuleRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&domain.UtmModule{}, id).Error
}

// DataTypes returns every module ordered by data_type. Each integration carries
// exactly one data_type, so the module catalog IS the data-type catalog (this
// replaces the removed utm_data_types table consumed by the rule editor).
func (r *pgModuleRepository) DataTypes(ctx context.Context) ([]domain.UtmModule, error) {
	var modules []domain.UtmModule
	err := r.db.WithContext(ctx).
		Model(&domain.UtmModule{}).
		Where("data_type <> ''").
		Order("data_type ASC").
		Find(&modules).Error
	if err != nil {
		return nil, err
	}
	return modules, nil
}

func (r *pgModuleRepository) Categories(ctx context.Context) ([]string, error) {
	var categories []string
	err := r.db.WithContext(ctx).
		Model(&domain.UtmModule{}).
		Distinct("module_category").
		Where("module_category <> ''").
		Order("module_category ASC").
		Pluck("module_category", &categories).Error
	if err != nil {
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
