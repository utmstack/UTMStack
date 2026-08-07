package repository

import (
	"context"
	"strings"
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/utils/tests"

	"github.com/utmstack/utmstack/backend/modules/datasources/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/database"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

const tenantA = "8f1c1b8e-0000-4000-8000-000000000001"

// newMultiTenantDB matches what modules.go wires when the licence is MSSP:
// tenancy callbacks are registered and Enabled() reports true, so scoped reads
// with no tenant must fail.
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

// The tenancy callback must add `tenant_id = ?` to reads on the datasources
// table when the caller carries a tenant. If it doesn't, the module leaks
// every tenant's rows to every other.
func TestCallbackScopesReadsByTenant(t *testing.T) {
	db := newMultiTenantDB(t)
	ctx := authz.WithTenantID(context.Background(), tenantA)

	stmt := db.Session(&gorm.Session{DryRun: true}).WithContext(ctx).
		Model(&domain.Datasource{}).
		Find(&[]domain.Datasource{}).Statement

	if !strings.Contains(stmt.SQL.String(), "tenant_id") {
		t.Fatalf("no tenant_id predicate in %q", stmt.SQL.String())
	}
}

// A read with no tenant on a multi-tenant instance must fail rather than span
// every tenant. This is the guard rail that catches missed handlers.
func TestReadWithoutTenantFails(t *testing.T) {
	db := newMultiTenantDB(t)
	stmt := db.Session(&gorm.Session{DryRun: true}).
		Model(&domain.Datasource{}).
		Find(&[]domain.Datasource{}).Statement

	if stmt.Error == nil {
		t.Fatal("a read with no tenant returned no error")
	}
}

// WithAllTenantsRead is what Enrichment uses to feed the alert plugin cache
// with every tenant's rows — the caller opts out of scoping, and the
// callback must respect it.
func TestWithAllTenantsReadSkipsScoping(t *testing.T) {
	db := newMultiTenantDB(t)
	ctx := tenancy.WithAllTenantsRead(authz.WithTenantID(context.Background(), tenantA))

	stmt := db.Session(&gorm.Session{DryRun: true}).WithContext(ctx).
		Model(&domain.Datasource{}).
		Where("group_id IS NOT NULL OR (labels IS NOT NULL AND labels <> '')").
		Find(&[]domain.Datasource{}).Statement

	if strings.Contains(stmt.SQL.String(), "tenant_id") {
		t.Fatalf("tenant_id predicate leaked into a WithAllTenantsRead query: %q", stmt.SQL.String())
	}
}

// The upsert paths (Ping and reconciler) run WithAllTenants because their
// batches span tenants and each row carries its own tenant. Assert that
// stamping still fires on WithAllTenants: rows arriving with an empty tenant
// are the caller's mistake and must not silently land unscoped.
func TestClearGroupScopesToTenant(t *testing.T) {
	db := newMultiTenantDB(t)
	provider := database.New(db)
	r := &pgDatasourceRepository{db: provider}

	ctx := authz.WithTenantID(context.Background(), tenantA)
	_ = r.ClearGroup(ctx, 42)

	stmt := db.Session(&gorm.Session{DryRun: true}).WithContext(ctx).
		Model(&domain.Datasource{}).
		Where("group_id = ?", 42).
		Update("group_id", nil).Statement

	sql := stmt.SQL.String()
	if !strings.Contains(sql, "tenant_id") {
		t.Fatalf("ClearGroup update missing tenant_id predicate: %q", sql)
	}
	if !strings.Contains(sql, "group_id") {
		t.Fatalf("ClearGroup update missing group_id predicate: %q", sql)
	}
}
