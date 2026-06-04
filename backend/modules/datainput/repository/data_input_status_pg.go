package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/utmstack/utmstack/backend/modules/datainput/connectors"
	"github.com/utmstack/utmstack/backend/modules/datainput/domain"
	"github.com/utmstack/utmstack/backend/modules/datainput/dto"
	"gorm.io/gorm"
)

const defaultMedian int64 = 86400

type pgDataInputStatusRepository struct {
	db *gorm.DB
}

func NewDataInputStatusRepository(db *gorm.DB) connectors.DataInputStatusRepository {
	return &pgDataInputStatusRepository{db: db}
}

func (r *pgDataInputStatusRepository) Create(ctx context.Context, status *domain.UtmDataInputStatus) error {
	status.ID = status.DataType + "-" + status.Source
	if status.Median == nil {
		m := defaultMedian
		status.Median = &m
	}
	return r.db.WithContext(ctx).Create(status).Error
}

func (r *pgDataInputStatusRepository) Update(ctx context.Context, status *domain.UtmDataInputStatus) error {
	return r.db.WithContext(ctx).Save(status).Error
}

func (r *pgDataInputStatusRepository) GetByID(ctx context.Context, id string) (*domain.UtmDataInputStatus, error) {
	var status domain.UtmDataInputStatus
	err := r.db.WithContext(ctx).First(&status, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &status, nil
}

func (r *pgDataInputStatusRepository) FindImportantDatasource(ctx context.Context, page, size int) ([]domain.UtmDataInputStatus, int64, error) {
	const baseQuery = `
SELECT * FROM utm_data_input_status d2
WHERE d2.source IN (
  SELECT d.source FROM utm_data_input_status d
  WHERE d.data_type NOT IN ('hids','utmstack','UTMStack')
  AND d.source IN (
    SELECT u2.source FROM utm_data_input_status u2
    WHERE u2.source NOT IN ('unknown')
    GROUP BY u2.source
  )
)`

	const countQuery = `
SELECT COUNT(*) FROM utm_data_input_status d2
WHERE d2.source IN (
  SELECT d.source FROM utm_data_input_status d
  WHERE d.data_type NOT IN ('hids','utmstack','UTMStack')
  AND d.source IN (
    SELECT u2.source FROM utm_data_input_status u2
    WHERE u2.source NOT IN ('unknown')
    GROUP BY u2.source
  )
)`

	var total int64
	if err := r.db.WithContext(ctx).Raw(countQuery).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := page * size
	paginatedQuery := baseQuery + fmt.Sprintf(" LIMIT %d OFFSET %d", size, offset)

	var items []domain.UtmDataInputStatus
	if err := r.db.WithContext(ctx).Raw(paginatedQuery).Scan(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *pgDataInputStatusRepository) Count(ctx context.Context, filters dto.DataInputStatusCountFilters) (int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.UtmDataInputStatus{})

	if filters.IDEquals != nil {
		q = q.Where("id = ?", *filters.IDEquals)
	}
	if filters.IDContains != nil {
		q = q.Where("id ILIKE ?", "%"+*filters.IDContains+"%")
	}
	if filters.SourceEquals != nil {
		q = q.Where("source = ?", *filters.SourceEquals)
	}
	if filters.SourceContains != nil {
		q = q.Where("source ILIKE ?", "%"+*filters.SourceContains+"%")
	}
	if filters.DataTypeEquals != nil {
		q = q.Where("data_type = ?", *filters.DataTypeEquals)
	}
	if filters.DataTypeContains != nil {
		q = q.Where("data_type ILIKE ?", "%"+*filters.DataTypeContains+"%")
	}
	if filters.TimestampEquals != nil {
		q = q.Where("timestamp = ?", *filters.TimestampEquals)
	}
	if filters.TimestampGreaterThan != nil {
		q = q.Where("timestamp > ?", *filters.TimestampGreaterThan)
	}
	if filters.TimestampLessThan != nil {
		q = q.Where("timestamp < ?", *filters.TimestampLessThan)
	}
	if filters.MedianEquals != nil {
		q = q.Where("median = ?", *filters.MedianEquals)
	}
	if filters.MedianGreaterThan != nil {
		q = q.Where("median > ?", *filters.MedianGreaterThan)
	}
	if filters.MedianLessThan != nil {
		q = q.Where("median < ?", *filters.MedianLessThan)
	}

	var count int64
	return count, q.Count(&count).Error
}

func (r *pgDataInputStatusRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.UtmDataInputStatus{}, "id = ?", id).Error
}

func (r *pgDataInputStatusRepository) FindDataSourcesToConfigure(ctx context.Context, excludeType string) ([]string, error) {
	const query = `
SELECT DISTINCT LOWER(TRIM(ds.data_type))
FROM utm_data_input_status ds
WHERE LOWER(TRIM(ds.data_type)) NOT IN (
  SELECT LOWER(TRIM(dt.data_type)) FROM utm_data_types dt
)
AND LOWER(TRIM(ds.data_type)) != LOWER(TRIM(?))
`
	var dataTypes []string
	err := r.db.WithContext(ctx).Raw(query, excludeType).Scan(&dataTypes).Error
	if err != nil {
		return []string{}, err
	}
	if dataTypes == nil {
		dataTypes = []string{}
	}
	return dataTypes, nil
}

func (r *pgDataInputStatusRepository) GetCheckpoint(ctx context.Context) (*domain.UtmDataInputStatusCheckpoint, error) {
	var cp domain.UtmDataInputStatusCheckpoint
	err := r.db.WithContext(ctx).First(&cp, 1).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cp, nil
}

func (r *pgDataInputStatusRepository) SaveCheckpoint(ctx context.Context, checkpoint *domain.UtmDataInputStatusCheckpoint) error {
	checkpoint.ID = 1
	return r.db.WithContext(ctx).Save(checkpoint).Error
}
