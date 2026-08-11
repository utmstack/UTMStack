package repository

import "github.com/utmstack/utmstack/backend/pkg/authz"

const (
	// Rules
	// The release's own copies, baked into the backend image. Overridable so a
	// developer can point at the definitions repo instead — see the equivalents
	// in soar (SOAR_FLOWS_SRC_DIR) and compliance (COMPLIANCE_SRC_DIR).
	RulesSrcDirEnv           = "EVENT_PROCESSING_RULES_SRC_DIR"
	DefaultSystemRulesSrcDir = "/utmstack/rules"
	RulesDirEnv              = "EVENT_PROCESSING_RULES_DIR"
	DefaultRulesDir          = "/workdir/rules"
	RuleFileExt              = ".yaml"

	// Pipelines
	PipelinesSrcDirEnv           = "EVENT_PROCESSING_PIPELINES_SRC_DIR"
	DefaultSystemPipelinesSrcDir = "/utmstack/filters"
	PipelinesDirEnv              = "EVENT_PROCESSING_PIPELINES_DIR"
	DefaultPipelinesDir          = "/workdir/pipeline/filters"
	PipelineFileExt              = ".yaml"

	// Pipeline dir (tenants.yaml + patterns.yaml live here, read by the engine).
	PipelineDirEnv     = "EVENT_PROCESSING_PIPELINE_DIR"
	DefaultPipelineDir = "/workdir/pipeline"
	TenantFileName     = "tenants.yaml"
	PatternsFileName   = "patterns.yaml"

	// Tenant UUID is a system-wide constant used by all plugins.
	DefaultTenantID   = authz.DefaultTenantID
	DefaultTenantName = "Default"

	// Shared
	SystemSubdir   = "system"
	UserSubdir     = "user"
	DisabledSuffix = ".disabled"
)
