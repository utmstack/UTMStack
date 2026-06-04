package domain

import "time"

type UtmDataInputStatus struct {
	ID        string  `gorm:"column:id;primaryKey"          json:"id"`
	Source    string  `gorm:"column:source;size:256;not null" json:"source"`
	DataType  string  `gorm:"column:data_type;size:50;not null" json:"dataType"`
	Timestamp int64   `gorm:"column:timestamp;not null"      json:"timestamp"` // epoch seconds
	Median    *int64  `gorm:"column:median"                  json:"median"`
	Alias     *string `gorm:"column:alias;size:500"          json:"alias"`
}

func (UtmDataInputStatus) TableName() string { return "utm_data_input_status" }

func (d *UtmDataInputStatus) IsDown() bool {
	if d.Timestamp == 0 || d.Median == nil {
		return false
	}
	elapsed := float64(time.Now().Unix() - d.Timestamp)
	return elapsed > float64(*d.Median)*1.5
}
