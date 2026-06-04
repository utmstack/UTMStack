package dto

import "github.com/utmstack/utmstack/backend/pkg/common_models"

type SearchRequest struct {
	Filters         []common_models.FilterType `json:"filters"`
	Top             int                        `json:"top"`
	IndexPattern    string                     `json:"indexPattern"`
	IncludeChildren bool                       `json:"includeChildren"`
}

type GenericSearchRequest struct {
	Index   string                     `json:"index"`
	Filters []common_models.FilterType `json:"filters"`
	Top     int                        `json:"top"`
}

type PropertyValuesRequest struct {
	Keyword      string `form:"keyword" binding:"required"`
	IndexPattern string `form:"indexPattern" binding:"required"`
}

type PropertyValuesWithCountRequest struct {
	Filters      []common_models.FilterType `json:"filters"`
	Field        string                     `json:"field" binding:"required"`
	Index        string                     `json:"index" binding:"required"`
	Top          int                        `json:"top"`
	OrderByCount bool                       `json:"orderByCount"`
	SortAsc      bool                       `json:"sortAsc"`
}

type CsvExportingParams struct {
	Filters      []common_models.FilterType `json:"filters"`
	Top          int                        `json:"top"`
	IndexPattern string                     `json:"indexPattern"`
	Columns      []string                   `json:"columns"`
}
