package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/incident_response/domain"
	"github.com/utmstack/utmstack/backend/modules/incident_response/dto"
)

type IncidentHistoryWriter interface {
	WriteHistory(ctx context.Context, incidentID int64, actionType, action, detail, by string) error
}

type ActionRepository interface {
	Save(action *domain.UtmIncidentAction) error
	FindByID(id int64) (*domain.UtmIncidentAction, error)
	FindAll(f dto.ActionFilter) ([]domain.UtmIncidentAction, int64, error)
	Delete(id int64) error
}

type ActionCommandRepository interface {
	Save(cmd *domain.UtmIncidentActionCommand) error
	FindByID(id int64) (*domain.UtmIncidentActionCommand, error)
	FindAll(f dto.ActionCommandFilter) ([]domain.UtmIncidentActionCommand, int64, error)
	Delete(id int64) error
}

type JobRepository interface {
	Save(job *domain.UtmIncidentJob) error
	FindByID(id int64) (*domain.UtmIncidentJob, error)
	FindAll(f dto.JobFilter) ([]domain.UtmIncidentJob, int64, error)
	Count(f dto.JobFilter) (int64, error)
	Delete(id int64) error
}

type VariableRepository interface {
	Save(v *domain.UtmIncidentVariable) error
	FindByID(id int64) (*domain.UtmIncidentVariable, error)
	FindAll(f dto.VariableFilter) ([]domain.UtmIncidentVariable, int64, error)
	FindAllPlain() ([]domain.UtmIncidentVariable, error)
	FindByName(name string) (*domain.UtmIncidentVariable, error)
	FindByNames(names []string) ([]domain.UtmIncidentVariable, error)
	Delete(id int64) error
}
