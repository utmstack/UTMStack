package repository

import (
	"context"
	"errors"
	"time"

	"github.com/utmstack/utmstack/backend/modules/logstash/connectors"
	"github.com/utmstack/utmstack/backend/modules/logstash/domain"
	"github.com/utmstack/utmstack/backend/modules/logstash/dto"
	"gorm.io/gorm"
)

type pgFilterRepository struct {
	db *gorm.DB
}

func NewFilterRepository(db *gorm.DB) connectors.FilterRepository {
	return &pgFilterRepository{db: db}
}

func (r *pgFilterRepository) Create(ctx context.Context, filter *domain.UtmLogstashFilter) error {
	now := time.Now().UTC()
	filter.UpdatedAt = &now
	return r.db.WithContext(ctx).Create(filter).Error
}

func (r *pgFilterRepository) Update(ctx context.Context, filter *domain.UtmLogstashFilter) error {
	now := time.Now().UTC()
	filter.UpdatedAt = &now
	return r.db.WithContext(ctx).Save(filter).Error
}

func (r *pgFilterRepository) GetByID(ctx context.Context, id int64) (*domain.UtmLogstashFilter, error) {
	var f domain.UtmLogstashFilter
	err := r.db.WithContext(ctx).First(&f, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &f, nil
}

func (r *pgFilterRepository) List(ctx context.Context, filters dto.FilterFilters) ([]domain.UtmLogstashFilter, int64, error) {
	q := r.applyWhere(r.db.WithContext(ctx).Model(&domain.UtmLogstashFilter{}), filters)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []domain.UtmLogstashFilter
	offset := filters.Page * filters.Size
	if err := q.Offset(offset).Limit(filters.Size).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *pgFilterRepository) applyWhere(q *gorm.DB, f dto.FilterFilters) *gorm.DB {
	if f.IDEquals != nil {
		q = q.Where("id = ?", *f.IDEquals)
	}
	if f.FilterNameContains != nil {
		q = q.Where("filter_name ILIKE ?", "%"+*f.FilterNameContains+"%")
	}
	if f.FilterGroupIDEq != nil {
		q = q.Where("filter_group_id = ?", *f.FilterGroupIDEq)
	}
	if f.FilterGroupIDGte != nil {
		q = q.Where("filter_group_id >= ?", *f.FilterGroupIDGte)
	}
	if f.FilterGroupIDLte != nil {
		q = q.Where("filter_group_id <= ?", *f.FilterGroupIDLte)
	}
	if f.IsActiveEq != nil {
		q = q.Where("is_active = ?", *f.IsActiveEq)
	}
	return q
}

func (r *pgFilterRepository) FiltersByPipelineID(ctx context.Context, pipelineID int64) ([]domain.UtmLogstashFilter, error) {
	var items []domain.UtmLogstashFilter
	sql := `SELECT ulf.* FROM
		(SELECT DISTINCT filter_id FROM utm_group_logstash_pipeline_filters WHERE pipeline_id = ?) bypipeline
		INNER JOIN utm_logstash_filter ulf ON bypipeline.filter_id = ulf.id`
	if err := r.db.WithContext(ctx).Raw(sql, pipelineID).Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *pgFilterRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&domain.UtmLogstashFilter{}, "id = ?", id).Error
}

func (r *pgFilterRepository) FindAllBySystemOwner(ctx context.Context) ([]domain.UtmLogstashFilter, error) {
	var items []domain.UtmLogstashFilter
	if err := r.db.WithContext(ctx).Where("system_owner = ?", true).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *pgFilterRepository) FindByModuleName(ctx context.Context, moduleName string) ([]domain.UtmLogstashFilter, error) {
	var items []domain.UtmLogstashFilter
	if err := r.db.WithContext(ctx).Where("module_name = ?", moduleName).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *pgFilterRepository) FindByFilterName(ctx context.Context, name string) (*domain.UtmLogstashFilter, error) {
	var f domain.UtmLogstashFilter
	err := r.db.WithContext(ctx).
		Where("filter_name = ?", name).
		First(&f).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &f, nil
}

func (r *pgFilterRepository) FindByLogstashFilterAndSystemOwner(ctx context.Context, content string) (*domain.UtmLogstashFilter, error) {
	var f domain.UtmLogstashFilter
	err := r.db.WithContext(ctx).
		Where("logstash_filter = ? AND system_owner = ?", content, true).
		First(&f).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &f, nil
}

// FindDataTypeByID resolves the data_type string from utm_data_types for the given id.
func (r *pgFilterRepository) FindDataTypeByID(ctx context.Context, dataTypeID int64) (string, error) {
	var dataType string
	err := r.db.WithContext(ctx).
		Raw("SELECT data_type FROM utm_data_types WHERE id = ?", dataTypeID).
		Scan(&dataType).Error
	if err != nil {
		return "", err
	}
	return dataType, nil
}
