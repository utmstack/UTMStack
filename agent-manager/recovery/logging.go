package recovery

import (
	"time"

	"github.com/threatwinds/go-sdk/catcher"
)

// All helpers in this file emit structured events via catcher.Info / catcher.Error,
// matching the project's existing logging pattern (see agent/agent_imp.go for style).
// Every event carries process="agent-manager" and an event= field for log filtering.

const logProcess = "agent-manager"

// LogRecoveryImported logs when a YAML is successfully imported into the DB for
// the first time (event: recovery_imported).
func LogRecoveryImported(yamlID, hash string, targetCount int) {
	catcher.Info("Recovery imported", map[string]any{
		"event":        "recovery_imported",
		"yaml_id":      yamlID,
		"yaml_hash":    hash,
		"target_count": targetCount,
		"process":      logProcess,
	})
}

// LogRecoverySkippedInvalid logs a per-file skip due to parse, schema, or
// sentinel validation failure (event: recovery_skipped_invalid).
func LogRecoverySkippedInvalid(path, reason string, err error) {
	catcher.Error("Recovery skipped (invalid)", err, map[string]any{
		"event":   "recovery_skipped_invalid",
		"path":    path,
		"reason":  reason,
		"process": logProcess,
	})
}

// LogRecoveryHashDrift logs when a YAML file's SHA-256 hash differs from the
// stored hash for the same yaml_id (event: recovery_hash_drift).
// The OLD content is kept; this is a warning that the file changed in place.
func LogRecoveryHashDrift(yamlID, dbHash, fsHash string) {
	catcher.Error("Recovery hash drift detected — keeping existing DB content", nil, map[string]any{
		"event":   "recovery_hash_drift",
		"yaml_id": yamlID,
		"db_hash": dbHash,
		"fs_hash": fsHash,
		"process": logProcess,
	})
}

// LogRecoveryBlocked logs when a recovery or target transitions to BLOCKED
// (event: recovery_blocked).
func LogRecoveryBlocked(yamlID, reasonKind, reasonDetail string) {
	catcher.Info("Recovery blocked", map[string]any{
		"event":         "recovery_blocked",
		"yaml_id":       yamlID,
		"reason_kind":   reasonKind,
		"reason_detail": reasonDetail,
		"process":       logProcess,
	})
}

// LogRecoveryUnblocked logs when a recovery transitions out of BLOCKED
// (event: recovery_unblocked).
func LogRecoveryUnblocked(yamlID string) {
	catcher.Info("Recovery unblocked", map[string]any{
		"event":   "recovery_unblocked",
		"yaml_id": yamlID,
		"process": logProcess,
	})
}

// LogRecoveryArchived logs when a recovery is marked ARCHIVED because its YAML
// file is no longer present on disk (event: recovery_archived).
func LogRecoveryArchived(yamlID string) {
	catcher.Info("Recovery archived (YAML removed from disk)", map[string]any{
		"event":   "recovery_archived",
		"yaml_id": yamlID,
		"process": logProcess,
	})
}

// LogRecoveryReactivated logs when a recovery is transitioned from BLOCKED or
// ARCHIVED back to ACTIVE (event: recovery_reactivated).
func LogRecoveryReactivated(yamlID string, hashDrift bool) {
	catcher.Info("Recovery reactivated", map[string]any{
		"event":      "recovery_reactivated",
		"yaml_id":    yamlID,
		"hash_drift": hashDrift,
		"process":    logProcess,
	})
}

// LogRecoveryPurged logs when a recovery and all its targets are hard-deleted
// by an operator CLI action (event: recovery_purged).
func LogRecoveryPurged(yamlID string, targetCount int) {
	catcher.Info("Recovery purged by operator", map[string]any{
		"event":           "recovery_purged",
		"yaml_id":         yamlID,
		"target_count":    targetCount,
		"operator_action": true,
		"process":         logProcess,
	})
}

// LogRecoveryDispatched logs after a successful stream.Send for a recovery
// script (event: recovery_dispatched).
func LogRecoveryDispatched(yamlID string, agentID uint, cmdID string, attempt int) {
	catcher.Info("Recovery dispatched to agent", map[string]any{
		"event":    "recovery_dispatched",
		"yaml_id":  yamlID,
		"agent_id": agentID,
		"cmd_id":   cmdID,
		"attempt":  attempt,
		"process":  logProcess,
	})
}

// LogRecoverySucceeded logs when the success sentinel is found in the agent's
// command result (event: recovery_succeeded).
func LogRecoverySucceeded(yamlID string, agentID uint, cmdID string) {
	catcher.Info("Recovery succeeded", map[string]any{
		"event":    "recovery_succeeded",
		"yaml_id":  yamlID,
		"agent_id": agentID,
		"cmd_id":   cmdID,
		"process":  logProcess,
	})
}

// LogRecoveryFailed logs when sentinel is absent or a send error occurred
// (event: recovery_failed). terminal=true when max_attempts is reached.
func LogRecoveryFailed(yamlID string, agentID uint, cmdID string, errMsg string, terminal bool) {
	catcher.Error("Recovery failed", nil, map[string]any{
		"event":    "recovery_failed",
		"yaml_id":  yamlID,
		"agent_id": agentID,
		"cmd_id":   cmdID,
		"error":    errMsg,
		"terminal": terminal,
		"process":  logProcess,
	})
}

// LogRecoveryRetryScheduled logs when a failed target is re-armed to PENDING
// for the next tick (event: recovery_retry_scheduled).
func LogRecoveryRetryScheduled(yamlID string, agentID uint, retryAt time.Time, attempt int) {
	catcher.Info("Recovery retry scheduled", map[string]any{
		"event":    "recovery_retry_scheduled",
		"yaml_id":  yamlID,
		"agent_id": agentID,
		"retry_at": retryAt.UTC().Format(time.RFC3339),
		"attempt":  attempt,
		"process":  logProcess,
	})
}

// LogRecoveryDependencyWait logs when dispatch is deferred because the
// dependency recovery has not yet succeeded for this agent
// (event: recovery_dependency_wait).
func LogRecoveryDependencyWait(yamlID string, agentID uint, blockedBy string) {
	catcher.Info("Recovery dispatch deferred — waiting for dependency", map[string]any{
		"event":      "recovery_dependency_wait",
		"yaml_id":    yamlID,
		"agent_id":   agentID,
		"blocked_by": blockedBy,
		"process":    logProcess,
	})
}

// LogRecoveryExpired logs when a recovery or its targets transition to EXPIRED
// because now() > expires_at (event: recovery_expired).
func LogRecoveryExpired(yamlID string) {
	catcher.Info("Recovery expired", map[string]any{
		"event":   "recovery_expired",
		"yaml_id": yamlID,
		"process": logProcess,
	})
}

// LogRecoveryDryRun logs when a dry-run dispatch is simulated for an agent
// without actually sending the script (event: recovery_dry_run).
func LogRecoveryDryRun(yamlID string, agentID uint) {
	catcher.Info("Recovery dry-run (no script sent)", map[string]any{
		"event":    "recovery_dry_run",
		"yaml_id":  yamlID,
		"agent_id": agentID,
		"process":  logProcess,
	})
}

// LogRecoverySkipped logs when an operator marks a target as SKIPPED via the
// CLI skip subcommand (event: recovery_skipped).
func LogRecoverySkipped(yamlID string, agentID uint) {
	catcher.Info("Recovery target skipped by operator", map[string]any{
		"event":           "recovery_skipped",
		"yaml_id":         yamlID,
		"agent_id":        agentID,
		"operator_action": true,
		"process":         logProcess,
	})
}

// LogRecoveryShutdownRollback logs the number of DISPATCHED targets rolled back
// to PENDING during graceful shutdown (event: recovery_shutdown_rollback).
func LogRecoveryShutdownRollback(count int) {
	catcher.Info("Recovery shutdown rollback: DISPATCHED→PENDING", map[string]any{
		"event":       "recovery_shutdown_rollback",
		"rolled_back": count,
		"process":     logProcess,
	})
}
