package dto

import "github.com/utmstack/utmstack/backend/pkg/database"

type VisualizationFilter struct {
	Name      string  `form:"name"`
	ChartType string  `form:"chartType"`
	IDPattern *uint64 `form:"idPattern"`
	database.Params
}
