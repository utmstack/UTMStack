package dto

import "github.com/utmstack/utmstack/backend/pkg/database"

type VisualizationFilter struct {
	Name string `form:"name"`
	database.Params
}
