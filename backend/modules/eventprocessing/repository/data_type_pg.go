package repository

import (
	"context"
	"errors"
	"time"

	"github.com/utmstack/utmstack/backend/modules/correlation/connectors"
	"github.com/utmstack/utmstack/backend/modules/correlation/domain"
	correrrors "github.com/utmstack/utmstack/backend/modules/correlation/errors"
	"github.com/utmstack/utmstack/backend/pkg/logger"
	"gorm.io/gorm"
)

type pgDataTypeRepository struct {
	db *gorm.DB
}

func NewDataTypeRepository(db *gorm.DB) connectors.DataTypeRepository {
	return &pgDataTypeRepository{db: db}
}

func (r *pgDataTypeRepository) Create(ctx context.Context, dt *domain.UtmDataTypes) (*domain.UtmDataTypes, error) {
	now := time.Now().UTC()
	dt.LastUpdate = &now
	if err := r.db.WithContext(ctx).Create(dt).Error; err != nil {
		return nil, err
	}
	return dt, nil
}

func (r *pgDataTypeRepository) Update(ctx context.Context, dt *domain.UtmDataTypes) (*domain.UtmDataTypes, error) {
	now := time.Now().UTC()
	dt.LastUpdate = &now
	if err := r.db.WithContext(ctx).Save(dt).Error; err != nil {
		return nil, err
	}
	return dt, nil
}

func (r *pgDataTypeRepository) GetByID(ctx context.Context, id int64) (*domain.UtmDataTypes, error) {
	var dt domain.UtmDataTypes
	err := r.db.WithContext(ctx).First(&dt, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &dt, nil
}

func (r *pgDataTypeRepository) List(ctx context.Context, f connectors.DataTypeFilters) ([]domain.UtmDataTypes, int64, error) {
	page, size := normPage(f.Page, f.Size)

	q := r.db.WithContext(ctx).Model(&domain.UtmDataTypes{})

	if f.Search != "" {
		like := "%" + f.Search + "%"
		q = q.Where("data_type ILIKE ? OR data_type_name ILIKE ? OR data_type_description ILIKE ?", like, like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []domain.UtmDataTypes
	if err := q.Order("id ASC").
		Offset(page * size).
		Limit(size).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *pgDataTypeRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&domain.UtmDataTypes{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return correrrors.ErrDataTypeNotFound
	}
	return nil
}

func (r *pgDataTypeRepository) FindDataSourcesToConfigure(ctx context.Context, excludeType string) ([]string, error) {
	const query = `
		SELECT DISTINCT LOWER(TRIM(di.data_type))
		FROM utm_data_input_status di
		WHERE LOWER(TRIM(di.data_type)) NOT IN (
			SELECT LOWER(TRIM(dt.data_type)) FROM utm_data_types dt
		)
		AND LOWER(TRIM(di.data_type)) != LOWER(TRIM($1))
	`
	var dataTypes []string
	err := r.db.WithContext(ctx).Raw(query, excludeType).Scan(&dataTypes).Error
	if err != nil {
		logger.Warn("data-type-sync: FindDataSourcesToConfigure failed — skipping new-type detection: " + err.Error())
		return []string{}, nil
	}
	if dataTypes == nil {
		dataTypes = []string{}
	}
	return dataTypes, nil
}

func (r *pgDataTypeRepository) FindOrphanDataSourceConfigurations(ctx context.Context) ([]domain.UtmDataTypes, error) {
	const query = `
		SELECT * FROM utm_data_types dt
		WHERE dt.system_owner = false
		AND dt.data_type NOT IN (SELECT DISTINCT di.source FROM utm_data_input_status di)
	`
	var items []domain.UtmDataTypes
	err := r.db.WithContext(ctx).Raw(query).Scan(&items).Error
	if err != nil {
		logger.Warn("data-type-sync: utm_data_input_status not available — skipping orphan detection: " + err.Error())
		return []domain.UtmDataTypes{}, nil
	}
	if items == nil {
		items = []domain.UtmDataTypes{}
	}
	return items, nil
}

func (r *pgDataTypeRepository) SaveAll(ctx context.Context, items []domain.UtmDataTypes) error {
	if len(items) == 0 {
		return nil
	}
	now := time.Now().UTC()
	for i := range items {
		items[i].LastUpdate = &now
	}
	return r.db.WithContext(ctx).Save(&items).Error
}

func (r *pgDataTypeRepository) FindByDataType(ctx context.Context, dataType string) (*domain.UtmDataTypes, error) {
	var dt domain.UtmDataTypes
	err := r.db.WithContext(ctx).
		Where("LOWER(data_type) = LOWER(?)", dataType).
		First(&dt).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &dt, nil
}
