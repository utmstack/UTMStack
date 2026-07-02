package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"github.com/utmstack/utmstack/backend/modules/compliance/dto"
)

type ScheduleRepository interface {
	Create(ctx context.Context, s *domain.UtmComplianceReportSchedule) error
	Update(ctx context.Context, s *domain.UtmComplianceReportSchedule) error
	GetByID(ctx context.Context, id int64) (*domain.UtmComplianceReportSchedule, error)
	ListByUser(ctx context.Context, userID int64, f dto.ScheduleFilters) ([]domain.UtmComplianceReportSchedule, int64, error)
	ListAll(ctx context.Context) ([]domain.UtmComplianceReportSchedule, error)
	Delete(ctx context.Context, id int64) error
}

type OpenSearchSQL interface {
	RunCheck(ctx context.Context, sql string) (hits int64, err error)
}

type OpenSearchAlerts interface {
	CountByRuleNames(ctx context.Context, ruleNames []string, sinceISO string) (int64, error)
}

type ReportStore interface {
	Save(ctx context.Context, snap *domain.ReportSnapshot) error
	List(ctx context.Context, frameworkKey string, limit int) ([]domain.ReportSnapshotMeta, error)
	Get(ctx context.Context, id string) (*domain.ReportSnapshot, error)
	Delete(ctx context.Context, id string) error
}

// ControlStatusOverrideRepository stores manual (framework, control) → status
// overrides. Upsert on the unique (framework_key, control_id) pair; ListByFramework
// returns a controlID → status map for the evaluator to consume.
type ControlStatusOverrideRepository interface {
	Upsert(ctx context.Context, o *domain.UtmComplianceControlStatusOverride) error
	Delete(ctx context.Context, frameworkKey, controlID string) error
	ListByFramework(ctx context.Context, frameworkKey string) (map[string]string, error)
}
