package connectors

import (
	"context"
	"github.com/google/uuid"
	"time"

	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"github.com/utmstack/utmstack/backend/modules/compliance/dto"
)

type ScheduleRepository interface {
	Create(ctx context.Context, s *domain.UtmComplianceReportSchedule) error
	Update(ctx context.Context, s *domain.UtmComplianceReportSchedule) error
	GetByID(ctx context.Context, id int64) (*domain.UtmComplianceReportSchedule, error)
	ListByUser(ctx context.Context, userID uuid.UUID, f dto.ScheduleFilters) ([]domain.UtmComplianceReportSchedule, int64, error)
	ListAll(ctx context.Context) ([]domain.UtmComplianceReportSchedule, error)
	Delete(ctx context.Context, id int64) error
	// ClaimDue atomically advances LastExecutionDate from expectedLast to
	// newLast, and reports whether THIS call won the claim. Guards against N
	// horizontally-scaled replicas all polling the same due schedule and
	// sending duplicate report emails: only the replica whose compare-and-swap
	// succeeds proceeds to delivery.
	ClaimDue(ctx context.Context, id int64, expectedLast, newLast time.Time) (bool, error)
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

// ControlNoteRepository stores freeform notes per (framework, control). Same
// upsert-on-unique pattern; ListByFramework returns a controlID → note map for
// the evaluator to attach to report rows.
type ControlNoteRepository interface {
	Upsert(ctx context.Context, n *domain.UtmComplianceControlNote) error
	Delete(ctx context.Context, frameworkKey, controlID string) error
	ListByFramework(ctx context.Context, frameworkKey string) (map[string]string, error)
}

// TenantFrameworkRepository records which frameworks each tenant possesses.
// Row-present = enabled; disable is a delete. ListForTenant is the input to
// the tenant-scoped evaluate loop. ListAll spans every tenant and is what the
// scheduler's sweep would use.
type TenantFrameworkRepository interface {
	List(ctx context.Context) ([]string, error)                           // framework keys the acting tenant possesses
	ListForTenant(ctx context.Context, tenantID string) ([]string, error) // same, for a specific tenant
	Enable(ctx context.Context, frameworkKey string) error                // upsert (tenant, framework)
	Disable(ctx context.Context, frameworkKey string) error               // delete (tenant, framework)
	Has(ctx context.Context, frameworkKey string) (bool, error)
	ListTenants(ctx context.Context, frameworkKey string) ([]string, error) // tenants who possess this framework
}
