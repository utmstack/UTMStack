package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/pkg/database"
)

type CreateVariableRequest struct {
	Name        string  `json:"variableName"        binding:"required"`
	Value       string  `json:"variableValue"       binding:"required"`
	Description *string `json:"variableDescription"`
	IsSecret    bool    `json:"isSecret"`
}

type UpdateVariableRequest struct {
	ID          uuid.UUID `json:"id"                  binding:"required"`
	Name        *string   `json:"variableName"`
	Value       *string   `json:"variableValue"`
	Description *string   `json:"variableDescription"`
	IsSecret    bool      `json:"isSecret"`
}

type VariableResponse struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"variableName"`
	Description  *string    `json:"variableDescription"`
	Value        string     `json:"variableValue"`
	IsSecret     bool       `json:"isSecret"`
	CreatedBy    string     `json:"createdBy"`
	CreatedAt    time.Time  `json:"createdAt"`
	ModifiedBy   string     `json:"lastModifiedBy"`
	ModifiedDate *time.Time `json:"lastModifiedDate"`
}

type VariableFilter struct {
	database.Params
	Name *string `form:"variableName"`
}
