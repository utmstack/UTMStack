package connectors

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"github.com/utmstack/utmstack/backend/modules/compliance/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type ScheduleRepository interface {
	Create(ctx context.Context, s *domain.ReportSchedule) error
	Update(ctx context.Context, s *domain.ReportSchedule) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ReportSchedule, error)
	ListByUser(ctx context.Context, userID uuid.UUID, f dto.ScheduleFilters) ([]domain.ReportSchedule, int64, error)
	ListDue(ctx context.Context, now time.Time) ([]domain.ReportSchedule, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ClaimDue(ctx context.Context, id uuid.UUID, expectedNext, newLast, newNext time.Time) (bool, error)
	ForFramework(ctx context.Context, frameworkKey string) (*domain.ReportSchedule, error)
}

type CheckQuery struct {
	Dataset  domain.Dataset
	DataType string
	Filters  []common_models.FilterType
	From, To time.Time
}

type EventCounter interface {
	Count(ctx context.Context, q CheckQuery) (int64, error)
	HasData(ctx context.Context, dataset domain.Dataset, dataType string, from, to time.Time) (bool, error)
	CountByRuleNames(ctx context.Context, ruleNames []string, from, to time.Time) (map[string]int64, error)
}

type ReportStore interface {
	Get(ctx context.Context, frameworkKey string) (*domain.Report, error)
	Save(ctx context.Context, r *domain.Report) error
	List(ctx context.Context) ([]domain.Report, error)
	Delete(ctx context.Context, frameworkKey string) error
}

type ReportScoreStore interface {
	Upsert(ctx context.Context, p *domain.ReportScore) error
	History(ctx context.Context, frameworkKey string, from, to time.Time) ([]domain.ReportScore, error)
	Body(ctx context.Context, frameworkKey string, day time.Time) ([]byte, error)
	PruneBodies(ctx context.Context, before time.Time) (int64, error)
}
