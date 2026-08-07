package connectors

import (
	"context"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/incidents/domain"
	"github.com/utmstack/utmstack/backend/modules/incidents/dto"
)

type IncidentRepository interface {
	// Create writes the incident together with the alerts it opens with. One
	// transaction, because an incident that lost half its alerts to a failed
	// insert is not a smaller incident, it is a wrong one — and the unique index
	// makes the obvious retry fail on the half that did land.
	Create(ctx context.Context, incident *domain.Incident, alerts []domain.IncidentAlert) error
	// LinkAlerts appends alerts to an existing incident and stores the severity
	// it ends up with, in the same transaction and for the same reason.
	LinkAlerts(ctx context.Context, incident *domain.Incident, alerts []domain.IncidentAlert) error
	Update(ctx context.Context, incident *domain.Incident) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Incident, error)
	FindAll(ctx context.Context, query dto.IncidentListQuery) ([]domain.Incident, int64, error)
	// DistinctAssignees is every assignee currently in use, for the filter to
	// offer. They are free text: a name the platform has never heard of is a
	// valid answer, because an incident can be handed to someone outside it.
	DistinctAssignees(ctx context.Context) ([]string, error)
}

type IncidentAlertRepository interface {
	Save(ctx context.Context, alert *domain.IncidentAlert) error
	Update(ctx context.Context, alert *domain.IncidentAlert) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.IncidentAlert, error)
	FindByIncidentID(ctx context.Context, incidentID uuid.UUID) ([]domain.IncidentAlert, error)
	// AlertIDsByIncident reads only the column the caller needs: propagating a
	// status to the incident's alerts wants their ids, not their rows.
	AlertIDsByIncident(ctx context.Context, incidentID uuid.UUID) ([]string, error)
	// WorstSeverity is the severity of the incident's worst remaining alert,
	// decided by the database. It runs on every link, edit and unlink, so it
	// returns one row rather than all of them.
	WorstSeverity(ctx context.Context, incidentID uuid.UUID) (domain.IncidentSeverity, error)
	FindByAlertIDs(ctx context.Context, alertIDs []string) ([]domain.IncidentAlert, error)
	FindAll(ctx context.Context, query dto.IncidentAlertListQuery) ([]domain.IncidentAlert, int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
	BulkUpdateStatus(ctx context.Context, alertIDs []string, status string) error
}

type IncidentNoteRepository interface {
	Save(ctx context.Context, note *domain.IncidentNote) error
	Update(ctx context.Context, note *domain.IncidentNote) error
	FindByIncidentID(ctx context.Context, incidentID uuid.UUID) ([]domain.IncidentNote, error)
	FindAll(ctx context.Context, query dto.IncidentNoteListQuery) ([]domain.IncidentNote, int64, error)
}

type IncidentHistoryRepository interface {
	Save(ctx context.Context, h *domain.IncidentHistory) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.IncidentHistory, error)
	FindAll(ctx context.Context, query dto.HistoryListQuery) ([]domain.IncidentHistory, int64, error)
	Count(ctx context.Context, query dto.HistoryListQuery) (int64, error)
}
