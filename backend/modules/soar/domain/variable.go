package domain

import "time"

type UtmIncidentVariable struct {
	ID                  int64      `gorm:"column:id;primaryKey;autoIncrement"  json:"id"`
	VariableName        *string    `gorm:"column:variable_name;uniqueIndex"    json:"variableName"`
	VariableValue       *string    `gorm:"column:variable_value"               json:"variableValue"`
	VariableDescription *string    `gorm:"column:variable_description"         json:"variableDescription"`
	IsSecret            bool       `gorm:"column:is_secret;not null"           json:"isSecret"`
	CreatedBy           *string    `gorm:"column:created_by"                   json:"createdBy"`
	LastModifiedDate    *time.Time `gorm:"column:last_modified_date"           json:"lastModifiedDate"`
	LastModifiedBy      *string    `gorm:"column:last_modified_by"             json:"lastModifiedBy"`
}

func (UtmIncidentVariable) TableName() string { return "utm_incident_variables" }
