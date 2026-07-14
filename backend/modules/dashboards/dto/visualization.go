package dto

import "github.com/utmstack/utmstack/backend/pkg/database"

type VisualizationFilter struct {
	DashboardID *uint64 `form:"dashboardId"`
	database.Params
}
