package domain

import "time"

type UtmDataTypes struct {
	ID                  int64      `gorm:"column:id;primaryKey"`
	DataType            string     `gorm:"column:data_type;size:250;not null"`
	DataTypeName        string     `gorm:"column:data_type_name;size:250;not null"`
	DataTypeDescription string     `gorm:"column:data_type_description"`
	LastUpdate          *time.Time `gorm:"column:last_update"`
	Included            bool       `gorm:"column:included;not null"`
	SystemOwner         bool       `gorm:"column:system_owner;not null"`
}

func (UtmDataTypes) TableName() string { return "utm_data_types" }
