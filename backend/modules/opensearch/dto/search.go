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
	Columns      []DataColumn               `json:"columns"`
}

// DataColumn describes one CSV column, mirroring the legacy DataColumn shape so
// the frontend payload stays compatible: Label is the header, Field is the
// (possibly .keyword-suffixed / dotted) source path, Type drives value formatting.
type DataColumn struct {
	Label   string `json:"label"`
	Field   string `json:"field"`
	Type    string `json:"type"`
	Visible bool   `json:"visible"`
}
