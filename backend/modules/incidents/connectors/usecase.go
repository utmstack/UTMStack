package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/incidents/domain"
	"github.com/utmstack/utmstack/backend/modules/incidents/dto"
)

type IncidentUsecase interface {
	Create(ctx context.Context, userLogin string, req dto.CreateIncidentRequest) (*domain.UtmIncident, error)
	AddAlerts(ctx context.Context, userLogin string, req dto.AddAlertsRequest) (*domain.UtmIncident, error)
	ChangeStatus(ctx context.Context, userLogin string, req dto.ChangeStatusRequest) (*domain.UtmIncident, error)
	Assign(ctx context.Context, userLogin string, req dto.AssignRequest) (*domain.UtmIncident, error)
	List(ctx context.Context, query dto.IncidentListQuery) ([]domain.UtmIncident, int64, error)
	GetByID(ctx context.Context, id int64) (*domain.UtmIncident, error)
	GetUsersAssigned(ctx context.Context) ([]dto.UserAssignedDTO, error)
}

type IncidentAlertUsecase interface {
	Create(ctx context.Context, req dto.IncidentAlertRequest) (*domain.UtmIncidentAlert, error)
	UpdateStatus(ctx context.Context, userLogin string, req dto.UpdateAlertStatusRequest) error
	Update(ctx context.Context, req dto.UpdateIncidentAlertRequest) (*domain.UtmIncidentAlert, error)
	List(ctx context.Context, query dto.IncidentAlertListQuery) ([]domain.UtmIncidentAlert, int64, error)
	Delete(ctx context.Context, id int64) error
}

type IncidentNoteUsecase interface {
	Create(ctx context.Context, userLogin string, req dto.CreateNoteRequest) (*domain.UtmIncidentNote, error)
	List(ctx context.Context, query dto.IncidentNoteListQuery) ([]domain.UtmIncidentNote, int64, error)
}

type IncidentHistoryUsecase interface {
	List(ctx context.Context, query dto.HistoryListQuery) ([]domain.UtmIncidentHistory, int64, error)
	Count(ctx context.Context, query dto.HistoryListQuery) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.UtmIncidentHistory, error)
}
