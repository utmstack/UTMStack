package dto

import (
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
	IndexPattern  string                     `json:"indexPattern"`
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
