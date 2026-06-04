package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/incidents/domain"
	"github.com/utmstack/utmstack/backend/modules/incidents/dto"
)

type IncidentRepository interface {
	Save(ctx context.Context, incident *domain.UtmIncident) error
	Update(ctx context.Context, incident *domain.UtmIncident) error
	FindByID(ctx context.Context, id int64) (*domain.UtmIncident, error)
	FindAll(ctx context.Context, query dto.IncidentListQuery) ([]domain.UtmIncident, int64, error)
	Delete(ctx context.Context, id int64) error
}

type IncidentAlertRepository interface {
	Save(ctx context.Context, alert *domain.UtmIncidentAlert) error
	Update(ctx context.Context, alert *domain.UtmIncidentAlert) error
	FindByID(ctx context.Context, id int64) (*domain.UtmIncidentAlert, error)
	FindByIncidentID(ctx context.Context, incidentID int64) ([]domain.UtmIncidentAlert, error)
	FindByAlertIDs(ctx context.Context, alertIDs []string) ([]domain.UtmIncidentAlert, error)
	FindAll(ctx context.Context, query dto.IncidentAlertListQuery) ([]domain.UtmIncidentAlert, int64, error)
	Delete(ctx context.Context, id int64) error
	ExistsByAlertID(ctx context.Context, alertID string) (bool, error)
	BulkUpdateStatus(ctx context.Context, alertIDs []string, status int) error
}

type IncidentNoteRepository interface {
	Save(ctx context.Context, note *domain.UtmIncidentNote) error
	Update(ctx context.Context, note *domain.UtmIncidentNote) error
	FindByIncidentID(ctx context.Context, incidentID int64) ([]domain.UtmIncidentNote, error)
	FindAll(ctx context.Context, query dto.IncidentNoteListQuery) ([]domain.UtmIncidentNote, int64, error)
}

type IncidentHistoryRepository interface {
	Save(ctx context.Context, h *domain.UtmIncidentHistory) error
	FindByID(ctx context.Context, id int64) (*domain.UtmIncidentHistory, error)
	FindAll(ctx context.Context, query dto.HistoryListQuery) ([]domain.UtmIncidentHistory, int64, error)
	Count(ctx context.Context, query dto.HistoryListQuery) (int64, error)
}
