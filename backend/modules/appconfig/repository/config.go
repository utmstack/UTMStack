package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/utmstack/utmstack/backend/modules/appconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/appconfig/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

type pgRepo struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) connectors.Repository {
	return &pgRepo{db: db}
}

var defaultTenant = uuid.MustParse(authz.DefaultTenantID)

func actingTenant(ctx context.Context) uuid.UUID {
	if t, err := uuid.Parse(authz.TenantIDFromContext(ctx)); err == nil {
		return t
	}
	return defaultTenant
}

func (r *pgRepo) inherited(ctx context.Context, tenant uuid.UUID) *gorm.DB {
	q := r.db.WithContext(tenancy.WithAllTenantsRead(ctx))
	if tenant == defaultTenant {
		return q.Where("tenant_id = ?", tenant)
	}
	return q.Where("tenant_id IN ?", []uuid.UUID{tenant, defaultTenant})
}

func preferOwn(rows []domain.Config, tenant uuid.UUID) *domain.Config {
	var inherited *domain.Config
	for i := range rows {
		if rows[i].TenantID == tenant {
			return &rows[i]
		}
		inherited = &rows[i]
	}
	return inherited
}

func (r *pgRepo) List(ctx context.Context) ([]domain.Config, error) {
	tenant := actingTenant(ctx)

	var rows []domain.Config
	if err := r.inherited(ctx, tenant).Order("key ASC").Find(&rows).Error; err != nil {
		return nil, err
	}

	byKey := make(map[string][]domain.Config, len(rows))
	keys := make([]string, 0, len(rows))
	for _, row := range rows {
		if _, seen := byKey[row.Key]; !seen {
			keys = append(keys, row.Key)
		}
		byKey[row.Key] = append(byKey[row.Key], row)
	}

	items := make([]domain.Config, 0, len(keys))
	for _, key := range keys {
		if row := preferOwn(byKey[key], tenant); row != nil {
			items = append(items, *row)
		}
	}
	return items, nil
}

func (r *pgRepo) GetByKey(ctx context.Context, key string) (*domain.Config, error) {
	tenant := actingTenant(ctx)

	var rows []domain.Config
	if err := r.inherited(ctx, tenant).Where("key = ?", key).Find(&rows).Error; err != nil {
		return nil, err
	}
	return preferOwn(rows, tenant), nil
}

func (r *pgRepo) GetOwn(ctx context.Context, key string) (*domain.Config, error) {
	var c domain.Config
	err := r.db.WithContext(tenancy.WithAllTenantsRead(ctx)).
		Where("tenant_id = ? AND key = ?", actingTenant(ctx), key).
		Take(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *pgRepo) Save(ctx context.Context, c *domain.Config) error {
	tenant := actingTenant(ctx)

	own, err := r.GetOwn(ctx, c.Key)
	if err != nil {
		return err
	}

	row := *c
	row.TenantID = tenant
	if own != nil {
		row.ID = own.ID
		return r.db.WithContext(ctx).Save(&row).Error
	}

	row.ID = uuid.Nil
	return r.db.WithContext(ctx).Create(&row).Error
}
