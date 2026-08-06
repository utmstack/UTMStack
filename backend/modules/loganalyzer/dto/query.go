package dto

import (
	"encoding/json"
	"time"

	"github.com/utmstack/utmstack/backend/pkg/common_models"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

type QueryFilter struct {
	Name  string `form:"name"`
	Owner string `form:"owner"`
	database.Params
}

type TopValue struct {
	Value   string  `json:"value"`
	Count   int64   `json:"count"`
	Percent float64 `json:"percent"`
}

type TopValuesResponse struct {
	Total int64      `json:"total"`
	Top   []TopValue `json:"top"`
}

type ChartViewRequest struct {
	Dataset       string                     `json:"dataset"`
	DataType      string                     `json:"dataType,omitempty"`
	Filters       []common_models.FilterType `json:"filters"`
	Interval      string                     `json:"interval"`
	Field         string                     `json:"field"`
	Top           int                        `json:"top"`
	FieldDataType string                     `json:"fieldDataType"`
}

type ChartViewResponse struct {
	Categories []string `json:"categories"`
	Values     []int64  `json:"values"`
}

type Field struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Filterable bool   `json:"filterable"`
	Searchable bool   `json:"searchable"`
}

// SearchRequest is the explorer asking for documents: a dataset, the filters on
// screen, the window, and where in the results to look.
type SearchRequest struct {
	Dataset  string                     `json:"dataset"`
	DataType string                     `json:"dataType,omitempty"`
	Filters  []common_models.FilterType `json:"filters"`
	From     *time.Time                 `json:"from,omitempty"`
	To       *time.Time                 `json:"to,omitempty"`
	Page     int                        `json:"page"`
	Size     int                        `json:"size"`
	SortBy   string                     `json:"sortBy,omitempty"`
	Order    string                     `json:"order,omitempty"`
}

type SearchResponse struct {
	Data  []json.RawMessage `json:"data"`
	Total int64             `json:"total"`
}
