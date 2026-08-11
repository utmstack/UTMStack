package connectors

import (
	"context"
	"encoding/json"
	"time"

	"github.com/utmstack/utmstack/backend/modules/eventprocessing/domain"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
)

// IngestionStatsQuery is one question about ingestion.
//
// The statistics dataset is not events: the stats plugin writes one row per
// (tenant, topic, source, data type) every cycle, carrying how many events and
// how many bytes that cycle saw. Every answer is therefore a sum of those
// counters — counting the rows would report how often the plugin wrote, which
// is a number that looks plausible and means nothing.
type IngestionStatsQuery struct {
	From, To time.Time
	// Type is the pipeline topic: enqueue_success, parsing_dropped,
	// analysis_dropped, correlation_dropped. Empty means every topic.
	Type       string
	DataSource string
}

type IngestionStatsRepository interface {
	TotalsByField(ctx context.Context, field string, q IngestionStatsQuery, top int) ([]dto.IngestionStatsBucket, dto.IngestionTotals, error)
	Timeline(ctx context.Context, q IngestionStatsQuery, interval string) ([]dto.TimelinePoint, error)
	TimelineByField(ctx context.Context, field string, q IngestionStatsQuery, interval string, top int) ([]dto.TimelineSeries, error)
}

type RuleRepository interface {
	Load() error
	Watch(ctx context.Context)
	List(f RuleListFilter) ([]*domain.StoredRule, int)
	FindByName(name string) *domain.StoredRule
	FindByRelPath(relPath string) *domain.StoredRule
	AllRelPaths() []string
	DistinctValues(prop, value, tenantId string) []string
	ReadRuleBytes(relPath string) ([]byte, error)
	CorrelationsByName(name string) (json.RawMessage, bool)
	Create(rule domain.Rule, tenantId string) (*domain.StoredRule, error)
	Update(relPath string, rule domain.Rule) (*domain.StoredRule, error)
	Delete(relPath string) error
	SetEnabled(tenantId, relPath string, enabled bool) (bool, error)
}

type RuleListFilter struct {
	Page            int // 0-based
	Size            int
	Name            string // case-insensitive partial on rule name
	Search          string // case-insensitive partial on rule name (general text)
	Active          *bool  // enabled state
	SystemOwner     *bool  // system overlay membership
	TenantId        string
	Categories      []string               // any-of
	Adversaries     []domain.AdversaryType // any-of
	Techniques      []string               // any-of
	DataTypes       []string               // any-of (rule must reference at least one)
	Confidentiality []int                  // any-of
	Integrity       []int                  // any-of
	Availability    []int                  // any-of
	InitDate        time.Time              // Modified >= InitDate (when non-zero)
	EndDate         time.Time              // Modified <= EndDate (when non-zero)
}

type PipelineRepository interface {
	Load() error
	Watch(ctx context.Context)
	List(tenantId string) []domain.Pipeline
	GetByRelPath(relPath string) *domain.Pipeline
	Create(relPath string, content []byte, tenantId string) (*domain.Pipeline, error)
	Update(relPath string, content []byte) (*domain.Pipeline, error)
	Delete(relPath string) error
	SetEnabled(tenantId, relPath string, active bool) error
}

type EngineConfigRepository interface {
	WriteTenants(tenants []domain.TenantAssets) error
	SetRuleDisabled(tenantId, identity string, disabled bool) error
	SetPipelineDisabled(tenantId, identity string, disabled bool) error
	SetPipelineOrder(tenantId string, order []string) error
	DisabledRuleSet(tenantId string) map[string]bool
	DisabledPipelineSet(tenantId string) map[string]bool
	PipelineOrder(tenantId string) []string
	ReadPatterns() (map[string]string, error)
	BuiltInPatterns() map[string]string
	WritePatterns(patterns map[string]string) error
	PokeRuleReload(absPath string)
}
