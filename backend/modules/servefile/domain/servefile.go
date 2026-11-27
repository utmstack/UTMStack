package domain

import (
	"time"

	"github.com/google/uuid"
)

// ServedFile is the registry row for a file stored on disk, encrypted, by the
// servefile module. The row is tenant-scoped: the GORM tenancy callback
// (pkg/tenancy) stamps TenantID on insert and appends WHERE tenant_id on
// every query, so no manual tenant filter lives here.
type ServedFile struct {
	ID          uuid.UUID `gorm:"column:id;primaryKey;type:uuid" json:"id"`
	TenantID    uuid.UUID `gorm:"column:tenant_id;type:uuid" json:"-"`
	Filename    string    `gorm:"column:filename" json:"filename"`
	ContentType string    `gorm:"column:content_type" json:"content_type"`
	Size        int64     `gorm:"column:size" json:"size"`
	CreatedAt   time.Time `json:"created_at"`
}

func (ServedFile) TableName() string { return "served_files" }
