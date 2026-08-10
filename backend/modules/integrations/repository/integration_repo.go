package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/utmstack/utmstack/backend/modules/integrations/connectors"
	"github.com/utmstack/utmstack/backend/modules/integrations/domain"
)

type pgIntegrationRepository struct {
	db *gorm.DB
}

func NewIntegrationRepository(db *gorm.DB) connectors.IntegrationRepository {
	return &pgIntegrationRepository{db: db}
}

var _ connectors.IntegrationRepository = (*pgIntegrationRepository)(nil)

func (r *pgIntegrationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Integration, error) {
	var i domain.Integration
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&i).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrIntegrationNotFound
	}
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *pgIntegrationRepository) GetByName(ctx context.Context, name string) (*domain.Integration, error) {
	var i domain.Integration
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&i).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrIntegrationNotFound
	}
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *pgIntegrationRepository) List(ctx context.Context, filter connectors.IntegrationListFilter) ([]domain.Integration, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.Integration{})

	if filter.IngestType != nil {
		q = q.Where("ingest_type = ?", *filter.IngestType)
	}
	if filter.NameContains != nil && *filter.NameContains != "" {
		q = q.Where("name ILIKE ?", "%"+*filter.NameContains+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []domain.Integration
	err := q.
		Order("name ASC").
		Limit(filter.Limit()).
		Offset(filter.Offset()).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *pgIntegrationRepository) Save(ctx context.Context, integration *domain.Integration) error {
	return r.db.WithContext(ctx).Save(integration).Error
}

func (r *pgIntegrationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Integration{}, "id = ?", id).Error
}

func (r *pgIntegrationRepository) DataTypes(ctx context.Context) ([]domain.Integration, error) {
	var items []domain.Integration
	err := r.db.WithContext(ctx).
		Model(&domain.Integration{}).
		Where("data_type <> ''").
		Order("data_type ASC").
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}
