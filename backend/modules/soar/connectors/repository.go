package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
)

type RuleFilters struct {
	ID                  int64
	RuleName            string
	RuleActive          *bool
	AgentPlatform       string
	CreatedBy           string
	LastModifiedBy      string
	CreatedDateGTE      string
	CreatedDateLTE      string
	LastModifiedDateGTE string
	LastModifiedDateLTE string
	SystemOwner         *bool
	// Page is 0-based. Size is the page size.
	Page int
	Size int
}

type TemplateFilters struct {
	ID          int64
	Label       string
	Description string
	Command     string
	SystemOwner *bool
	Page        int
	Size        int
}

type ExecutionFilters struct {
	ID                       int64
	RuleID                   int64
	RuleIDGreaterThanOrEqual *int64
	RuleIDLessThanOrEqual    *int64
	AlertID                  string
	Agent                    string
	ExecutionStatus          domain.ExecutionStatus
	NonExecutionCause        domain.NonExecutionCause
	ExecutionDateGTE         string
	ExecutionDateLTE         string
	Page                     int
	Size                     int
}

type RuleRepository interface {
	Create(ctx context.Context, rule *domain.AlertResponseRule) (*domain.AlertResponseRule, error)
	Update(ctx context.Context, rule *domain.AlertResponseRule) (*domain.AlertResponseRule, error)
	GetByID(ctx context.Context, id int64) (*domain.AlertResponseRule, error)
	List(ctx context.Context, f RuleFilters) ([]domain.AlertResponseRule, int64, error)
	Delete(ctx context.Context, id int64) error
}

type TemplateRepository interface {
	List(ctx context.Context, f TemplateFilters) ([]domain.AlertResponseActionTemplate, int64, error)
}

type ExecutionRepository interface {
	Create(ctx context.Context, e *domain.AlertResponseRuleExecution) (*domain.AlertResponseRuleExecution, error)
	List(ctx context.Context, f ExecutionFilters) ([]domain.AlertResponseRuleExecution, int64, error)
}

type ResolveFilterRepository interface {
	GetAgentPlatforms(ctx context.Context) ([]string, error)
	GetUsers(ctx context.Context) ([]string, error)
}
