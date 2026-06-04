package connectors

import (
	"context"
	"time"

	incidents_connectors "github.com/utmstack/utmstack/backend/modules/incidents/connectors"
	incidents_domain "github.com/utmstack/utmstack/backend/modules/incidents/domain"
)

type incidentHistoryAdapter struct {
	repo incidents_connectors.IncidentHistoryRepository
}

func NewIncidentHistoryWriterAdapter(repo incidents_connectors.IncidentHistoryRepository) IncidentHistoryWriter {
	return &incidentHistoryAdapter{repo: repo}
}

func (a *incidentHistoryAdapter) WriteHistory(ctx context.Context, incidentID int64, actionType, action, detail, by string) error {
	h := &incidents_domain.UtmIncidentHistory{
		IncidentID:        incidentID,
		ActionType:        actionType,
		Action:            action,
		ActionDetail:      &detail,
		ActionCreatedDate: time.Now().UTC(),
		ActionCreatedBy:   &by,
	}
	return a.repo.Save(ctx, h)
}
