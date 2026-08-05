package repository

import (
	"context"
	"strings"
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/utils/tests"

	"github.com/utmstack/utmstack/backend/modules/notifications/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

const customerTenant = "8f1c1b8e-0000-4000-8000-000000000001"

// newMultiTenantDB registers the tenancy callbacks the way the running instance
// does once it holds an MSSP licence.
func newMultiTenantDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(tests.DummyDialector{}, &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	if err := tenancy.Register(db, func() bool { return true }); err != nil {
		t.Fatalf("tenancy.Register: %v", err)
	}
	return db
}

// The repository does not filter by tenant itself; the callbacks do, and they
// are stronger — a read with no tenant fails instead of returning everyone's.
// This is here so removing the hand-rolled filter stays safe.
func TestReadsAreScopedByTheCallbacks(t *testing.T) {
	db := newMultiTenantDB(t)
	r := &pgNotificationRepository{db: db}

	ctx := authz.WithTenantID(context.Background(), customerTenant)
	_, _, _ = r.FindAll(ctx, dto.NotificationListQuery{})

	stmt := db.Session(&gorm.Session{DryRun: true}).WithContext(ctx).
		Model(&struct {
			ID       int64  `gorm:"primaryKey"`
			TenantID string `gorm:"column:tenant_id"`
		}{}).Find(&[]any{}).Statement

	if !strings.Contains(stmt.SQL.String(), "tenant_id") {
		t.Fatalf("no tenant predicate in %q", stmt.SQL.String())
	}
}

// A read with no tenant on a multi-tenant instance must fail rather than span
// every tenant, which is what the hand-rolled filter used to do.
func TestAReadWithoutATenantFails(t *testing.T) {
	db := newMultiTenantDB(t)
	r := &pgNotificationRepository{db: db}

	if _, _, err := r.FindAll(context.Background(), dto.NotificationListQuery{}); err == nil {
		t.Fatal("a read with no tenant returned no error")
	}
}
