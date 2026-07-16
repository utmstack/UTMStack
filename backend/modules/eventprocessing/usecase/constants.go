package usecase

import "time"

const (
	// Rules
	SystemRulesSrcDir = "/utmstack/rules"
	RulesDirEnv       = "EVENT_PROCESSING_RULES_DIR"
	DefaultRulesDir   = "/workdir/rules"
	RuleFileExt       = ".yaml"

	// Filters
	SystemFiltersSrcDir = "/utmstack/filters"
	FiltersDirEnv       = "EVENT_PROCESSING_FILTERS_DIR"
	DefaultFiltersDir   = "/workdir/pipeline/filters"
	FilterFileExt       = ".yaml"

	// Pipeline dir (tenants.yaml + patterns.yaml live here, read by the engine).
	PipelineDirEnv     = "EVENT_PROCESSING_PIPELINE_DIR"
	DefaultPipelineDir = "/workdir/pipeline"
	TenantFileName     = "tenants.yaml"
	PatternsFileName   = "patterns.yaml"

	// Tenant UUID is a system-wide constant used by all plugins.
	DefaultTenantID   = "ce66672c-e36d-4761-a8c8-90058fee1a24"
	DefaultTenantName = "Default"

	// Shared
	SystemSubdir   = "system"
	UserSubdir     = "user"
	DisabledSuffix = ".disabled"

	// Playground
	MaxPlaygroundBodyBytes  int64         = 1 << 20 // 1 MiB
	SemaphoreAcquireTimeout time.Duration = 30 * time.Second
	PlaygroundUserFilename  string        = "playground-user.yaml"
)
