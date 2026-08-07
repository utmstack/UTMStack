package connectors

import (
	"context"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/incidents/domain"
	"github.com/utmstack/utmstack/backend/modules/incidents/dto"
)

type IncidentUsecase interface {
	Create(ctx context.Context, userEmail string, req dto.CreateIncidentRequest) (*domain.Incident, error)
	AddAlerts(ctx context.Context, userEmail string, req dto.AddAlertsRequest) (*domain.Incident, error)
	ChangeStatus(ctx context.Context, userEmail string, req dto.ChangeStatusRequest) (*domain.Incident, error)
	Assign(ctx context.Context, userEmail string, req dto.AssignRequest) (*domain.Incident, error)
	List(ctx context.Context, query dto.IncidentListQuery) ([]domain.Incident, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Incident, error)
	GetAssignees(ctx context.Context) ([]string, error)
}

type IncidentAlertUsecase interface {
	Create(ctx context.Context, req dto.IncidentAlertRequest) (*domain.IncidentAlert, error)
	UpdateStatus(ctx context.Context, userEmail string, req dto.UpdateAlertStatusRequest) error
	Update(ctx context.Context, req dto.UpdateIncidentAlertRequest) (*domain.IncidentAlert, error)
	List(ctx context.Context, query dto.IncidentAlertListQuery) ([]domain.IncidentAlert, int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type IncidentNoteUsecase interface {
	Create(ctx context.Context, userEmail string, req dto.CreateNoteRequest) (*domain.IncidentNote, error)
	List(ctx context.Context, query dto.IncidentNoteListQuery) ([]domain.IncidentNote, int64, error)
}

type IncidentHistoryUsecase interface {
	List(ctx context.Context, query dto.HistoryListQuery) ([]domain.IncidentHistory, int64, error)
	Count(ctx context.Context, query dto.HistoryListQuery) (int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.IncidentHistory, error)
}
