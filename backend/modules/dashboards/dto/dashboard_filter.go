
package dto

import (
	"github.com/google/uuid"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

type DashboardFilterFilter struct {
	DashboardID *uuid.UUID
	database.Params
}


// the wire form of the filter. gin cannot bind a
// uuid.UUID from a query string (it is a [16]byte), so the id arrives as text
// and is parsed here.
type DashboardFilterQuery struct {
	DashboardID string `form:"dashboardId"`
	database.Params
}

func (q DashboardFilterQuery) Filter() (DashboardFilterFilter, error) {
	f := DashboardFilterFilter{Params: q.Params}
	if q.DashboardID == "" {
		return f, nil
	}
	id, err := uuid.Parse(q.DashboardID)
	if err != nil {
		return f, err
	}
	f.DashboardID = &id
	return f, nil
}
