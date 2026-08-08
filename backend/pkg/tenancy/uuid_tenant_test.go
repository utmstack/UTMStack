package tenancy

import (
	"testing"

	"github.com/google/uuid"
)

// uuidScopedModel mirrors the compliance tables, whose TenantID is a uuid.UUID
// rather than a string. Stamping goes through gorm's field setter, which is
// where a type it could not convert would leave the column zeroed instead of
// erroring — every row landing on the nil tenant, and every read finding them.
type uuidScopedModel struct {
	ID       uint64    `gorm:"primaryKey"`
	TenantID uuid.UUID `gorm:"column:tenant_id;type:uuid"`
	Name     string
}

func (uuidScopedModel) TableName() string { return "uuid_scoped" }

func TestStampFillsUUIDTenantField(t *testing.T) {
	db := newDB(t, multi)
	tid := uuid.New()

	row := uuidScopedModel{Name: "x"}
	if err := db.WithContext(ctxTenant(tid.String())).Create(&row).Error; err != nil {
		t.Fatalf("Create: %v", err)
	}
	if row.TenantID != tid {
		t.Fatalf("TenantID = %v, want %v", row.TenantID, tid)
	}
}

func TestUUIDModelStillFailsClosedWithoutTenant(t *testing.T) {
	db := newDB(t, multi)
	var out []uuidScopedModel
	if err := db.WithContext(ctxTenant("")).Find(&out).Error; err == nil {
		t.Fatal("unscoped read on a uuid-tenant model returned no error")
	}
}
