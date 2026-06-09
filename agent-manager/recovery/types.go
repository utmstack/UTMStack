package recovery

// RecoveryStatus represents the lifecycle state of a Recovery record.
type RecoveryStatus string

const (
	RecoveryStatusActive    RecoveryStatus = "ACTIVE"
	RecoveryStatusBlocked   RecoveryStatus = "BLOCKED"
	RecoveryStatusArchived  RecoveryStatus = "ARCHIVED"
	RecoveryStatusPaused    RecoveryStatus = "PAUSED"
	RecoveryStatusCompleted RecoveryStatus = "COMPLETED"
	RecoveryStatusExpired   RecoveryStatus = "EXPIRED"
	RecoveryStatusDisabled  RecoveryStatus = "DISABLED"
)

// TargetStatus represents the execution state of a RecoveryTarget row.
type TargetStatus string

const (
	TargetStatusPending    TargetStatus = "PENDING"
	TargetStatusDispatched TargetStatus = "DISPATCHED"
	TargetStatusSucceeded  TargetStatus = "SUCCEEDED"
	TargetStatusFailed     TargetStatus = "FAILED"
	TargetStatusExpired    TargetStatus = "EXPIRED"
	TargetStatusSkipped    TargetStatus = "SKIPPED"
	TargetStatusBlocked    TargetStatus = "BLOCKED"
)

// BlockedReason kind constants — used as the prefix in recoveries.blocked_reason column.
// Format: "<kind>:<detail>", e.g. "dependency_missing:fix-updater-v11.2.6-windows-amd64".
const (
	BlockedReasonDependencyMissing = "dependency_missing"
	BlockedReasonCycleDetected     = "cycle_detected"
)

// DefaultSuccessPattern is the sentinel that must appear as a standalone
// stdout line for a target to be marked SUCCEEDED when success_pattern is
// not explicitly set. The double-underscore wrap is intentional: the token
// must be unique enough that it cannot appear inside an error message,
// stack trace, or log line without the author meaning to emit it.
const DefaultSuccessPattern = "__UTMSTACK_RECOVERY_OK__"

// RecoveryYAML is the in-memory representation of a parsed recovery YAML file.
// It maps 1:1 with the YAML schema on disk. Do NOT add GORM tags here;
// the GORM model lives in models/recovery.go.
type RecoveryYAML struct {
	YamlID         string     `yaml:"yaml_id"`
	Name           string     `yaml:"name"`
	Description    string     `yaml:"description"`
	Shell          string     `yaml:"shell"`
	Target         YAMLTarget `yaml:"target"`
	SuccessPattern string     `yaml:"success_pattern"`
	MaxConcurrency int        `yaml:"max_concurrency"`
	MaxAttempts    int        `yaml:"max_attempts"`
	RetryAfter     int        `yaml:"retry_after_seconds"`
	AckTimeout     int        `yaml:"ack_timeout_seconds"`
	ExpiresAt      string     `yaml:"expires_at"`
	DryRun         bool       `yaml:"dry_run"`
	DependsOn      string     `yaml:"depends_on"`
	Script         string     `yaml:"script"`
	// Hash is the SHA-256 hex of the source file's content, populated by
	// yaml_loader.ParseAll after unmarshalling. Excluded from YAML parsing.
	Hash string `yaml:"-"`
}

// YAMLTarget is the nested `target:` block inside a RecoveryYAML.
//
// NOTE: architecture filtering is intentionally NOT supported in v1.
// The agent reports OS type and platform/distribution but does NOT report
// CPU architecture (amd64/arm64). Adding it requires a proto + agent change.
// For v1, recoveries are filtered by os + version only. Authors writing a
// recovery for a specific arch should encode that limitation in the script
// itself (e.g. early-exit if $env:PROCESSOR_ARCHITECTURE -ne "AMD64").
type YAMLTarget struct {
	OS         string `yaml:"os"`
	VersionLte string `yaml:"version_lte"`
}
