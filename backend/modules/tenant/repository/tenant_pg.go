package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/utmstack/utmstack/backend/modules/tenant/connectors"
	"github.com/utmstack/utmstack/backend/modules/tenant/domain"
	"github.com/utmstack/utmstack/backend/modules/tenant/dto"
)

type pgTenantRepository struct{ db *gorm.DB }

func NewTenantRepository(db *gorm.DB) connectors.TenantRepository {
	return &pgTenantRepository{db: db}
}

func (r *pgTenantRepository) Create(ctx context.Context, t *domain.Tenant) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *pgTenantRepository) Update(ctx context.Context, t *domain.Tenant) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *pgTenantRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.Tenant{}).Error
}

func (r *pgTenantRepository) FindByID(ctx context.Context, id string) (*domain.Tenant, error) {
	return r.findOne(ctx, "id = ?", id)
}

func (r *pgTenantRepository) FindByDomain(ctx context.Context, host string) (*domain.Tenant, error) {
	return r.findOne(ctx, "domain = ?", host)
}

func (r *pgTenantRepository) findOne(ctx context.Context, query string, arg any) (*domain.Tenant, error) {
	var t domain.Tenant
	err := r.db.WithContext(ctx).Where(query, arg).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *pgTenantRepository) List(ctx context.Context, f dto.Filter) ([]domain.Tenant, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.Tenant{})

	if f.Name != "" {
		q = q.Where("name ILIKE ?", "%"+f.Name+"%")
	}
	if f.Domain != "" {
		q = q.Where("domain ILIKE ?", "%"+f.Domain+"%")
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []domain.Tenant
	if err := q.Order("created_at DESC").
		Offset(f.Page * f.Size).
		Limit(f.Size).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}
