package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/modules/alerts/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type AdversaryUsecase interface {
	FetchAdversaryAlerts(ctx context.Context, filters []common_models.FilterType) ([]dto.AdversaryResponse, error)
}

type AlertUsecase interface {
	UpdateStatus(ctx context.Context, userLogin string, req dto.UpdateAlertStatusRequest) error
	UpdateNotes(ctx context.Context, userLogin string, alertID string, notes string) error
	UpdateTags(ctx context.Context, userLogin string, req dto.UpdateAlertTagsRequest) error
	ConvertToIncident(ctx context.Context, userLogin string, req dto.ConvertToIncidentRequest) error
	CountOpenAlerts(ctx context.Context) (*dto.CountOpenAlertsResponse, error)
}

type AlertTagUsecase interface {
	Create(ctx context.Context, req dto.CreateAlertTagRequest) (*domain.UtmAlertTag, error)
	Update(ctx context.Context, req dto.UpdateAlertTagRequest) (*domain.UtmAlertTag, error)
	List(ctx context.Context, filters dto.AlertTagFilters) ([]domain.UtmAlertTag, int64, error)
	GetByID(ctx context.Context, id uint64) (*domain.UtmAlertTag, error)
	Delete(ctx context.Context, id uint64) error
}

type AlertTagRuleUsecase interface {
	Create(ctx context.Context, req dto.CreateAlertTagRuleRequest) (*dto.AlertTagRuleResponse, error)
	Update(ctx context.Context, req dto.UpdateAlertTagRuleRequest) (*dto.AlertTagRuleResponse, error)
	List(ctx context.Context, filters dto.AlertTagRuleFilters) ([]dto.AlertTagRuleResponse, int64, error)
	GetByID(ctx context.Context, id uint64) (*dto.AlertTagRuleResponse, error)
	GetByIDs(ctx context.Context, ids []int64) ([]dto.AlertTagRuleResponse, error)
	Delete(ctx context.Context, id uint64) error
}
