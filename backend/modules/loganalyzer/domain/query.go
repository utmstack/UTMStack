package domain

import "time"

type SavedQuery struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TenantID    string    `gorm:"column:tenant_id;size:36;index" json:"-"`
	Name        string    `gorm:"column:name;size:100;not null" json:"name"`
	Description string    `gorm:"column:description" json:"description"`
	Owner       string    `gorm:"column:owner" json:"owner"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updatedAt"`
	Columns     string    `gorm:"column:columns" json:"columns"`
	Filters     string    `gorm:"column:filters" json:"filters"`
	Dataset     string    `gorm:"column:dataset" json:"dataset"`
}

func (SavedQuery) TableName() string { return "saved_query" }
