package tenancy

import (
	"context"
	"errors"
	"testing"
)

// A batch that legitimately spans tenants — one plugin watching the whole
// instance and flushing whatever changed — must keep the tenant each row
// carries. Stamping the acting tenant over them would file one tenant's users
// under another, and nothing would report it.
func TestSpanningLeavesEachRowsTenantAlone(t *testing.T) {
	db := newDB(t, multi)

	rows := []scopedModel{
		{Name: "from-a", TenantID: "tenant-a"},
		{Name: "from-b", TenantID: "tenant-b"},
	}
	err := db.WithContext(WithAllTenants(context.Background())).Create(&rows).Error
	if err != nil {
		t.Fatalf("spanning create = %v, want no error", err)
	}

	if rows[0].TenantID != "tenant-a" || rows[1].TenantID != "tenant-b" {
		t.Fatalf("tenants = %q/%q, want them untouched", rows[0].TenantID, rows[1].TenantID)
	}
}

// And the contrast that makes it necessary: acting for one tenant stamps the
// whole batch, which is right for a user's request and wrong for an ingest.
func TestActingForOneTenantStampsTheWholeBatch(t *testing.T) {
	db := newDB(t, multi)

	rows := []scopedModel{
		{Name: "from-a", TenantID: "tenant-a"},
		{Name: "from-b", TenantID: "tenant-b"},
	}
	db.WithContext(ctxTenant("tenant-a")).Create(&rows)

	if rows[1].TenantID != "tenant-a" {
		t.Fatalf("rows[1].TenantID = %q, want it stamped with the acting tenant", rows[1].TenantID)
	}
}

// An audit entry for platform work has no tenant to be stamped with —
// provisioning a tenant, a plugin posting a notification. Refusing it for that
// reason would drop exactly the records most worth keeping, and the caller
// swallows the error, so nothing would say so.
func TestSpanningAllowsARowWithNoTenant(t *testing.T) {
	db := newDB(t, multi)

	row := scopedModel{Name: "platform-audit"}
	if err := db.WithContext(WithAllTenants(context.Background())).Create(&row).Error; err != nil {
		t.Fatalf("spanning create with no tenant = %v, want no error", err)
	}
	if row.TenantID != "" {
		t.Fatalf("TenantID = %q, want it left empty", row.TenantID)
	}
}

// And without spanning it is refused, which is the behaviour that was losing
// those entries.
func TestWritingWithNoTenantIsRefusedWithoutSpanning(t *testing.T) {
	db := newDB(t, multi)

	row := scopedModel{Name: "platform-audit"}
	if err := db.WithContext(context.Background()).Create(&row).Error; !errors.Is(err, ErrNoTenant) {
		t.Fatalf("create = %v, want ErrNoTenant", err)
	}
}
