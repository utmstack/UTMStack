package domain

import (
	"time"

	"github.com/google/uuid"
)


type DashboardFilterType = string

const (
	DashboardFilterMultiple   DashboardFilterType = "multiple"
	DashboardFilterSearchable DashboardFilterType = "searchable"
)

type DashboardFilter struct {
	ID           uuid.UUID            `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID     uuid.UUID            `gorm:"column:tenant_id;type:uuid;not null;index" json:"-"`
	DashboardID  uuid.UUID            `gorm:"column:dashboard_id;type:uuid;not null;index" json:"dashboardId"`
	DataSet      string               `gorm:"column:data_set;not null" json:"data_set"`
	Field        string               `gorm:"column:field;not null" json:"field"`
	Type         DashboardFilterType  `gorm:"column:type;not null" json:"type"`
	Label        string               `gorm:"column:label" json:"label"`
	PlaceHolder  string               `gorm:"column:place_holder" json:"place_holder"`
	CreatedDate  time.Time            `gorm:"column:created_date" json:"createdDate"`
	ModifiedDate time.Time            `gorm:"column:modified_date" json:"modifiedDate"`
}

func (DashboardFilter) TableName() string { return "dashboard_filter" }

