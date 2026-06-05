package dto

import "time"

type CreateVariableRequest struct {
	VariableName        string  `json:"variableName"        binding:"required"`
	VariableDescription *string `json:"variableDescription"`
	VariableValue       *string `json:"variableValue"`
	IsSecret            bool    `json:"isSecret"`
}

type UpdateVariableRequest struct {
	ID                  int64   `json:"id"                  binding:"required"`
	VariableName        *string `json:"variableName"`
	VariableDescription *string `json:"variableDescription"`
	VariableValue       *string `json:"variableValue"`
	IsSecret            bool    `json:"isSecret"`
}

type VariableResponse struct {
	ID                  int64      `json:"id"`
	VariableName        *string    `json:"variableName"`
	VariableDescription *string    `json:"variableDescription"`
	VariableValue       *string    `json:"variableValue"`
	IsSecret            bool       `json:"isSecret"`
	CreatedBy           *string    `json:"createdBy"`
	LastModifiedDate    *time.Time `json:"lastModifiedDate"`
	LastModifiedBy      *string    `json:"lastModifiedBy"`
}

type VariableFilter struct {
	Page         int     `form:"page"`
	Size         int     `form:"size"`
	VariableName *string `form:"variableName"`
}
