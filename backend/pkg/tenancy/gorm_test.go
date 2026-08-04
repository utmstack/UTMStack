package tenancy

import (
	"context"
	"errors"
	"strings"
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/utils/tests"

	"github.com/utmstack/utmstack/backend/pkg/authz"
)

// scopedModel is tenant-scoped: it has a TenantID field.
type scopedModel struct {
	ID       uint64 `gorm:"primaryKey"`
	TenantID string `gorm:"column:tenant_id"`
	Name     string
}

func (scopedModel) TableName() string { return "scoped" }

// systemModel is tenant-scoped and also holds rows created by the product.
type systemModel struct {
	ID          uint64 `gorm:"primaryKey"`
	TenantID    string `gorm:"column:tenant_id"`
	SystemOwner bool   `gorm:"column:system_owner"`
	Name        string
}

func (systemModel) TableName() string        { return "systemish" }
func (systemModel) SystemFlagColumn() string { return "system_owner" }

// globalModel has no TenantID, so the callbacks must leave it alone.
type globalModel struct {
	ID   uint64 `gorm:"primaryKey"`
	Name string
}

func (globalModel) TableName() string { return "global" }

// newDB builds a gorm handle that renders SQL without ever connecting: these
// tests are about which predicate the callbacks add, not about the database.
func newDB(t *testing.T, enabled func() bool) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(tests.DummyDialector{}, &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	if err := Register(db, enabled); err != nil {
		t.Fatalf("Register: %v", err)
	}
	return db
}

func multi() bool  { return true }
func single() bool { return false }

func ctxTenant(id string) context.Context {
	return authz.WithTenantID(context.Background(), id)
}

func TestRegisterRejectsNilPredicate(t *testing.T) {
	db, err := gorm.Open(tests.DummyDialector{}, &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	if err := Register(db, nil); err == nil {
		t.Fatal("Register(nil) returned no error; a nil predicate would silently disable scoping")
	}
}

// The whole point of the package: a query with no tenant in context is an
// error, never an unfiltered read.
func TestReadWithoutTenantFailsClosed(t *testing.T) {
	db := newDB(t, multi)

	var out []scopedModel
	err := db.WithContext(context.Background()).Find(&out).Error

	if !errors.Is(err, ErrNoTenant) {
		t.Fatalf("Find without a tenant = %v, want ErrNoTenant", err)
	}
}

func TestWriteWithoutTenantFailsClosed(t *testing.T) {
	for _, tc := range []struct {
		name string
		run  func(*gorm.DB) error
	}{
		{"update", func(db *gorm.DB) error {
			return db.Model(&scopedModel{}).Where("id = ?", 1).Update("name", "x").Error
		}},
		{"delete", func(db *gorm.DB) error {
			return db.Where("id = ?", 1).Delete(&scopedModel{}).Error
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := newDB(t, multi)
			if err := tc.run(db.WithContext(context.Background())); !errors.Is(err, ErrNoTenant) {
				t.Fatalf("%s without a tenant = %v, want ErrNoTenant", tc.name, err)
			}
		})
	}
}

func TestReadIsScopedToTheTenant(t *testing.T) {
	db := newDB(t, multi)

	var out []scopedModel
	stmt := db.WithContext(ctxTenant("tenant-a")).Find(&out).Statement
	sql := stmt.SQL.String()

	if !strings.Contains(sql, "tenant_id") {
		t.Fatalf("no tenant predicate in %q", sql)
	}
	if !containsArg(stmt.Vars, "tenant-a") {
		t.Fatalf("tenant not bound as an argument; vars = %v", stmt.Vars)
	}
}

// System rows belong to no tenant and every tenant sees them.
func TestReadIncludesSystemRows(t *testing.T) {
	db := newDB(t, multi)

	var out []systemModel
	sql := db.WithContext(ctxTenant("tenant-a")).Find(&out).Statement.SQL.String()

	if !strings.Contains(sql, "system_owner") {
		t.Fatalf("system rows are excluded from reads: %q", sql)
	}
	if !strings.Contains(strings.ToUpper(sql), " OR ") {
		t.Fatalf("system rows should widen the predicate, not narrow it: %q", sql)
	}
}

// ...but a tenant must never be able to modify or delete one, which is what
// makes system rows immutable from inside an instance.
func TestWriteExcludesSystemRows(t *testing.T) {
	db := newDB(t, multi)

	sql := db.WithContext(ctxTenant("tenant-a")).
		Model(&systemModel{}).Where("id = ?", 1).Update("name", "x").
		Statement.SQL.String()

	if strings.Contains(sql, "system_owner") {
		t.Fatalf("a tenant can reach system rows on write: %q", sql)
	}
}

// A caller cannot create a row under another tenant by filling the field in.
func TestCreateStampsAndOverwritesTenant(t *testing.T) {
	db := newDB(t, multi)

	row := scopedModel{Name: "x", TenantID: "tenant-b"}
	db.WithContext(ctxTenant("tenant-a")).Create(&row)

	if row.TenantID != "tenant-a" {
		t.Fatalf("TenantID = %q, want it overwritten with the acting tenant", row.TenantID)
	}
}

func TestCreateStampsEveryRowOfASlice(t *testing.T) {
	db := newDB(t, multi)

	rows := []scopedModel{{Name: "a"}, {Name: "b", TenantID: "tenant-z"}}
	db.WithContext(ctxTenant("tenant-a")).Create(&rows)

	for i, r := range rows {
		if r.TenantID != "tenant-a" {
			t.Fatalf("rows[%d].TenantID = %q, want %q", i, r.TenantID, "tenant-a")
		}
	}
}

func TestCreateWithoutTenantFailsClosed(t *testing.T) {
	db := newDB(t, multi)

	row := scopedModel{Name: "x"}
	err := db.WithContext(context.Background()).Create(&row).Error

	if !errors.Is(err, ErrNoTenant) {
		t.Fatalf("Create without a tenant = %v, want ErrNoTenant", err)
	}
}

// WithAllTenants is how backfills and maintenance span the instance. It has to
// be explicit, and it must not error.
func TestWithAllTenantsSpansEveryTenant(t *testing.T) {
	db := newDB(t, multi)

	var out []scopedModel
	stmt := db.WithContext(WithAllTenants(context.Background())).Find(&out).Statement

	if err := stmt.Error; err != nil {
		t.Fatalf("WithAllTenants read = %v, want no error", err)
	}
	if containsArg(stmt.Vars, "tenant-a") {
		t.Fatalf("WithAllTenants still bound a tenant: %v", stmt.Vars)
	}
}

// A model with no TenantID is global; the callbacks must not touch it, with or
// without a tenant in context.
func TestModelWithoutTenantFieldIsUntouched(t *testing.T) {
	db := newDB(t, multi)

	var out []globalModel
	err := db.WithContext(context.Background()).Find(&out).Error

	if err != nil {
		t.Fatalf("global model read = %v, want no error", err)
	}
}

// On a single-tenant install every callback is a no-op, so nothing about it
// changes — including that a read with no tenant is perfectly fine.
func TestSingleTenantInstallIsInert(t *testing.T) {
	db := newDB(t, single)

	var out []scopedModel
	stmt := db.WithContext(context.Background()).Find(&out).Statement

	if err := stmt.Error; err != nil {
		t.Fatalf("single-tenant read = %v, want no error", err)
	}
	if strings.Contains(stmt.SQL.String(), "tenant_id") {
		t.Fatalf("single-tenant install got a tenant predicate: %q", stmt.SQL.String())
	}

	row := scopedModel{Name: "x", TenantID: "left-alone"}
	db.WithContext(context.Background()).Create(&row)
	if row.TenantID != "left-alone" {
		t.Fatalf("TenantID = %q, want it untouched on a single-tenant install", row.TenantID)
	}
}

// The predicate is consulted per operation, not once at Register: a licence can
// be activated while the process is running.
func TestPredicateIsConsultedPerOperation(t *testing.T) {
	on := false
	db := newDB(t, func() bool { return on })

	var out []scopedModel
	if err := db.WithContext(context.Background()).Find(&out).Error; err != nil {
		t.Fatalf("read while single-tenant = %v, want no error", err)
	}

	on = true
	if err := db.WithContext(context.Background()).Find(&out).Error; !errors.Is(err, ErrNoTenant) {
		t.Fatalf("read after the licence turned multi-tenant = %v, want ErrNoTenant", err)
	}
}

func containsArg(vars []any, want string) bool {
	for _, v := range vars {
		if s, ok := v.(string); ok && s == want {
			return true
		}
	}
	return false
}
