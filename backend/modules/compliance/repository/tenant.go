package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

var ErrNoTenant = errors.New("compliance: no tenant in scope")

func tenantFromCtx(ctx context.Context) uuid.UUID {
	tid, _ := uuid.Parse(authz.TenantIDFromContext(ctx))
	return tid
}

func scopeTenant(ctx context.Context, q *gorm.DB) *gorm.DB {
	if tid := tenantFromCtx(ctx); tid != uuid.Nil {
		return q.Where("tenant_id = ?", tid)
	}
	if tenancy.Enabled() {
		q.AddError(ErrNoTenant)
	}
	return q
}
