package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/correlation/domain"
	"gorm.io/gorm"
)

type RegexPatternFilters struct {
	Search string
	// Page is 0-based.
	Page int
	Size int
}

type TenantConfigFilters struct {
	Search string
	// Page is 0-based.
	Page int
	Size int
}

type RegexPatternRepository interface {
	Create(ctx context.Context, p *domain.UtmRegexPattern) (*domain.UtmRegexPattern, error)
	Update(ctx context.Context, p *domain.UtmRegexPattern) (*domain.UtmRegexPattern, error)
	GetByID(ctx context.Context, id int64) (*domain.UtmRegexPattern, error)
	List(ctx context.Context, f RegexPatternFilters) ([]domain.UtmRegexPattern, int64, error)
	Delete(ctx context.Context, id int64) error
}

type TenantConfigRepository interface {
	Create(ctx context.Context, t *domain.UtmTenantConfig) (*domain.UtmTenantConfig, error)
	Update(ctx context.Context, t *domain.UtmTenantConfig) (*domain.UtmTenantConfig, error)
	GetByID(ctx context.Context, id int64) (*domain.UtmTenantConfig, error)
	List(ctx context.Context, f TenantConfigFilters) ([]domain.UtmTenantConfig, int64, error)
	Delete(ctx context.Context, id int64) error
}

type DataTypeFilters struct {
	Search string
	// Page is 0-based.
	Page int
	Size int
}

type DataTypeRepository interface {
	Create(ctx context.Context, dt *domain.UtmDataTypes) (*domain.UtmDataTypes, error)
	Update(ctx context.Context, dt *domain.UtmDataTypes) (*domain.UtmDataTypes, error)
	GetByID(ctx context.Context, id int64) (*domain.UtmDataTypes, error)
	List(ctx context.Context, f DataTypeFilters) ([]domain.UtmDataTypes, int64, error)
	Delete(ctx context.Context, id int64) error

	// FindDataSourcesToConfigure returns sources in utm_data_input_status that
	// are not yet in utm_data_types, excluding the given type.
	// Returns empty slice (no error) if utm_data_input_status does not exist yet
	// — handled gracefully until module #13 is ported.
	FindDataSourcesToConfigure(ctx context.Context, excludeType string) ([]string, error)

	// FindOrphanDataSourceConfigurations returns non-system data types whose
	// data_type value no longer appears in utm_data_input_status.
	FindOrphanDataSourceConfigurations(ctx context.Context) ([]domain.UtmDataTypes, error)

	SaveAll(ctx context.Context, items []domain.UtmDataTypes) error

	FindByDataType(ctx context.Context, dataType string) (*domain.UtmDataTypes, error)
}

type CorrelationRuleFilters struct {
	// Page is 0-based.
	Page int
	Size int

	// Partial match on rule_name (ILIKE).
	RuleName string

	// Exact-match lists (IN semantics — nil means no filter).
	RuleActive          *bool
	RuleCategory        []string
	RuleAdversary       []string
	RuleTechnique       []string
	RuleConfidentiality []int
	RuleIntegrity       []int
	RuleAvailability    []int
	SystemOwner         *bool
	DataTypes           []string

	// Date-range filter on rule_last_update (ISO-8601 strings; inclusive).
	InitDate string
	EndDate  string

	// General text search against rule_name.
	Search string
}

type CorrelationRuleRepository interface {
	Create(ctx context.Context, rule *domain.UtmCorrelationRules) error
	Update(ctx context.Context, rule *domain.UtmCorrelationRules) error
	GetByID(ctx context.Context, id int64) (*domain.UtmCorrelationRules, error)
	List(ctx context.Context, filters CorrelationRuleFilters) ([]domain.UtmCorrelationRules, int64, error)
	Delete(ctx context.Context, id int64) error
	ActivateDeactivate(ctx context.Context, id int64, active bool) error

	FindDistinctPropertyValues(ctx context.Context, prop, value string) ([]string, error)

	SyncDataTypes(ctx context.Context, tx *gorm.DB, ruleID int64, dataTypes []domain.UtmDataTypes) error

	FindByRuleName(ctx context.Context, ruleName string) (*domain.UtmCorrelationRules, error)

	FindAllBySystemOwner(ctx context.Context, systemOwner bool) ([]domain.UtmCorrelationRules, error)
}
