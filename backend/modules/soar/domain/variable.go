package domain

import (
	"time"

	"github.com/google/uuid"
)

type SoarVariable struct {
	TenantID    uuid.UUID  `gorm:"column:tenant_id;type:uuid;not null;uniqueIndex:idx_soar_variable_tenant_name,priority:1" json:"-"`
	ID          uuid.UUID  `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"                                 json:"id"`
	Name        string     `gorm:"column:variable_name;not null;uniqueIndex:idx_soar_variable_tenant_name,priority:2"       json:"variableName"`
	Value       string     `gorm:"column:variable_value;not null;check:chk_soar_variable_value,variable_value <> ''"        json:"variableValue"`
	Description *string    `gorm:"column:variable_description"                                                              json:"variableDescription"`
	IsSecret    bool       `gorm:"column:is_secret;not null"                                                                json:"isSecret"`
	CreatedBy   string     `gorm:"column:created_by;size:150;not null"  json:"createdBy"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null"           json:"createdAt"`
	ModifiedBy  string     `gorm:"column:modified_by;size:150;not null" json:"lastModifiedBy"`
	ModifiedAt  *time.Time `gorm:"column:modified_at"                   json:"lastModifiedDate"`
}

func (SoarVariable) TableName() string { return "soar_variables" }
