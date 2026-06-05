package dto

import (
	"time"

	"github.com/utmstack/utmstack/backend/modules/correlation/domain"
)

type CreateDataTypeRequest struct {
	ID                  *int64 `json:"id"`
	DataType            string `json:"dataType"`
	DataTypeName        string `json:"dataTypeName"`
	DataTypeDescription string `json:"dataTypeDescription"`
	Included            bool   `json:"included"`
}

type UpdateDataTypeRequest struct {
	ID                  *int64 `json:"id"`
	DataType            string `json:"dataType"`
	DataTypeName        string `json:"dataTypeName"`
	DataTypeDescription string `json:"dataTypeDescription"`
	Included            bool   `json:"included"`
}

type DataTypeResponse struct {
	ID                  int64      `json:"id"`
	DataType            string     `json:"dataType"`
	DataTypeName        string     `json:"dataTypeName"`
	DataTypeDescription string     `json:"dataTypeDescription"`
	LastUpdate          *time.Time `json:"lastUpdate"`
	Included            bool       `json:"included"`
	SystemOwner         bool       `json:"systemOwner"`
}

type DataTypeFilters struct {
	// Page is 0-based (matches Java Spring Pageable).
	Page   int    `form:"page"`
	Size   int    `form:"size"`
	Search string `form:"search"`
}

type UpdateIncludeExcludeItem struct {
	ID       *int64 `json:"id"`
	Included bool   `json:"included"`
}

func DataTypeToResponse(e *domain.UtmDataTypes) *DataTypeResponse {
	return &DataTypeResponse{
		ID:                  e.ID,
		DataType:            e.DataType,
		DataTypeName:        e.DataTypeName,
		DataTypeDescription: e.DataTypeDescription,
		LastUpdate:          e.LastUpdate,
		Included:            e.Included,
		SystemOwner:         e.SystemOwner,
	}
}
