package repository

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

type pgTenantFrameworkRepo struct{ db *gorm.DB }

func NewTenantFrameworkRepository(db *gorm.DB) connectors.TenantFrameworkRepository {
	return &pgTenantFrameworkRepo{db: db}
}

func (r *pgTenantFrameworkRepo) List(ctx context.Context) ([]string, error) {
	return r.ListForTenant(ctx, authz.TenantIDFromContext(ctx))
}

func (r *pgTenantFrameworkRepo) ListForTenant(ctx context.Context, tenantID string) ([]string, error) {
	if tenantID == "" {
		return nil, nil
	}
	var keys []string
	if err := r.db.WithContext(ctx).
		Model(&domain.TenantFramework{}).
		Where("tenant_id = ?", tenantID).
		Order("framework_key").
		Pluck("framework_key", &keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}

func (r *pgTenantFrameworkRepo) Enable(ctx context.Context, frameworkKey string) error {
	tid := authz.TenantIDFromContext(ctx)
	if tid == "" || frameworkKey == "" {
		return domain.ErrInvalidID
	}
	// ON CONFLICT DO NOTHING keeps the CreatedAt of the first enable so
	// a repeat toggle doesn't rewrite when the framework was first taken on.
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).
		Create(&domain.TenantFramework{TenantID: tid, FrameworkKey: frameworkKey}).Error
}

func (r *pgTenantFrameworkRepo) Disable(ctx context.Context, frameworkKey string) error {
	tid := authz.TenantIDFromContext(ctx)
	if tid == "" || frameworkKey == "" {
		return domain.ErrInvalidID
	}
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND framework_key = ?", tid, frameworkKey).
		Delete(&domain.TenantFramework{}).Error
}

func (r *pgTenantFrameworkRepo) Has(ctx context.Context, frameworkKey string) (bool, error) {
	tid := authz.TenantIDFromContext(ctx)
	if tid == "" || frameworkKey == "" {
		return false, nil
	}
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&domain.TenantFramework{}).
		Where("tenant_id = ? AND framework_key = ?", tid, frameworkKey).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *pgTenantFrameworkRepo) ListTenants(ctx context.Context, frameworkKey string) ([]string, error) {
	if frameworkKey == "" {
		return nil, nil
	}
	var tids []string
	if err := r.db.WithContext(ctx).
		Model(&domain.TenantFramework{}).
		Where("framework_key = ?", frameworkKey).
		Order("tenant_id").
		Pluck("tenant_id", &tids).Error; err != nil {
		return nil, err
	}
	return tids, nil
}
