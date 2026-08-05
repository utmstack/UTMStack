package domain

import "time"

type UtmIncidentVariable struct {
	TenantID            string     `gorm:"column:tenant_id;size:36;not null;default:'';index;uniqueIndex:idx_incident_variable_tenant_name,priority:1" json:"-"`
	ID                  int64      `gorm:"column:id;primaryKey;autoIncrement"  json:"id"`
	VariableName        *string    `gorm:"column:variable_name;uniqueIndex:idx_incident_variable_tenant_name,priority:2" json:"variableName"`
	VariableValue       *string    `gorm:"column:variable_value"               json:"variableValue"`
	VariableDescription *string    `gorm:"column:variable_description"         json:"variableDescription"`
	IsSecret            bool       `gorm:"column:is_secret;not null"           json:"isSecret"`
	CreatedBy           *string    `gorm:"column:created_by"                   json:"createdBy"`
	LastModifiedDate    *time.Time `gorm:"column:last_modified_date"           json:"lastModifiedDate"`
	LastModifiedBy      *string    `gorm:"column:last_modified_by"             json:"lastModifiedBy"`
}

func (UtmIncidentVariable) TableName() string { return "utm_incident_variables" }
