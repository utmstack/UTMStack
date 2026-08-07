package repository

import (
	"context"

	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
	"gorm.io/gorm"
)

// tenantFromCtx returns the raw tenant-id string from ctx, or "" for on-prem/global actors.
func tenantFromCtx(ctx context.Context) string {
	return authz.TenantIDFromContext(ctx)
}

// scopeTenant narrows q to the acting tenant. Skipped when ctx carries
// tenancy.WithAllTenants (dispatcher cross-tenant scans).
func scopeTenant(ctx context.Context, q *gorm.DB) *gorm.DB {
	if tenancy.SpansAllTenants(ctx) {
		return q
	}
	if tid := tenantFromCtx(ctx); tid != "" {
		return q.Where("tenant_id = ?", tid)
	}
	return q
}
