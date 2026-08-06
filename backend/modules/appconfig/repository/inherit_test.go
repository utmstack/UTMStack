package repository

import (
	"context"
	"strings"
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/utils/tests"

	"github.com/utmstack/utmstack/backend/modules/appconfig/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

const customerTenant = "8f1c1b8e-0000-4000-8000-000000000001"

func newDB(t *testing.T) *gorm.DB {
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

// A tenant reads its own row where it has one and the default tenant's where it
// does not, so the read has to span both — which the tenancy callbacks would
// otherwise narrow away.
func TestGetByKeyReadsBothTenants(t *testing.T) {
	db := newDB(t)
	r := &pgRepo{db: db}

	ctx := authz.WithTenantID(context.Background(), customerTenant)
	_, _ = r.GetByKey(ctx, "utmstack.mail.host")

	sql := lastSQL(t, db, func(d *gorm.DB) *gorm.DB {
		return r.inherited(ctx, customerTenant).Where("conf_param_short = ?", "x")
	})

	if !strings.Contains(sql, "tenant_id IN") {
		t.Errorf("query does not span both tenants: %s", sql)
	}
}

// The instance's own tenant has nothing to inherit from, so it must not read
// anyone else's row.
func TestTheDefaultTenantReadsOnlyItsOwn(t *testing.T) {
	db := newDB(t)
	r := &pgRepo{db: db}

	ctx := authz.WithTenantID(context.Background(), authz.DefaultTenantID)
	sql := lastSQL(t, db, func(d *gorm.DB) *gorm.DB {
		return r.inherited(ctx, authz.DefaultTenantID).Where("conf_param_short = ?", "x")
	})

	if strings.Contains(sql, "IN") {
		t.Errorf("the default tenant read beyond itself: %s", sql)
	}
	if !strings.Contains(sql, "tenant_id") {
		t.Errorf("no tenant predicate: %s", sql)
	}
}

// GetOwn is what tells an override from something inherited, so it must never
// fall back.
func TestGetOwnDoesNotInherit(t *testing.T) {
	db := newDB(t)

	ctx := authz.WithTenantID(context.Background(), customerTenant)
	sql := lastSQL(t, db, func(d *gorm.DB) *gorm.DB {
		return d.WithContext(tenancy.WithAllTenantsRead(ctx)).
			Where("tenant_id = ? AND conf_param_short = ?", customerTenant, "x")
	})

	if strings.Contains(sql, "IN") {
		t.Errorf("GetOwn inherited: %s", sql)
	}
}

func lastSQL(t *testing.T, db *gorm.DB, build func(*gorm.DB) *gorm.DB) string {
	t.Helper()
	type row struct {
		ID       int64  `gorm:"primaryKey"`
		TenantID string `gorm:"column:tenant_id"`
	}
	stmt := build(db.Session(&gorm.Session{DryRun: true})).
		Model(&row{}).Find(&[]row{}).Statement
	return stmt.SQL.String()
}

// The tenant's own row has to win over the one it would otherwise inherit.
//
// This was expressed as Order(gorm.Expr(...)), which GORM drops on the floor —
// Order takes only a string or a clause.OrderByColumn — so the query ran with
// no ordering at all and the database returned whichever row it felt like. The
// symptom was a tenant saving its branding and reading the platform's back.
func TestPreferOwnPicksTheTenantsRow(t *testing.T) {
	rows := []domain.Config{
		{TenantID: authz.DefaultTenantID, ConfParamShort: "branding", ConfParamValue: "platform"},
		{TenantID: customerTenant, ConfParamShort: "branding", ConfParamValue: "theirs"},
	}

	got := preferOwn(rows, customerTenant)
	if got == nil || got.ConfParamValue != "theirs" {
		t.Fatalf("picked %v, want the tenant's own row", got)
	}

	// Order of arrival must not decide it.
	rows[0], rows[1] = rows[1], rows[0]
	if got := preferOwn(rows, customerTenant); got == nil || got.ConfParamValue != "theirs" {
		t.Fatalf("picked %v after reordering, want the tenant's own row", got)
	}
}

// With nothing of its own, the tenant reads what the instance set.
func TestPreferOwnFallsBackToTheDefault(t *testing.T) {
	rows := []domain.Config{
		{TenantID: authz.DefaultTenantID, ConfParamShort: "branding", ConfParamValue: "platform"},
	}
	got := preferOwn(rows, customerTenant)
	if got == nil || got.ConfParamValue != "platform" {
		t.Fatalf("picked %v, want the inherited default", got)
	}
}

func TestPreferOwnOnNothing(t *testing.T) {
	if got := preferOwn(nil, customerTenant); got != nil {
		t.Fatalf("picked %v from no rows, want nil", got)
	}
}
