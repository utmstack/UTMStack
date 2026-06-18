package recovery

import (
	"fmt"
	"time"

	"github.com/utmstack/UTMStack/agent-manager/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TargetWithRecovery is a target row plus its parent recovery, used by the
// dispatcher to know dispatch parameters without N+1 queries.
type TargetWithRecovery struct {
	Target   models.RecoveryTarget
	Recovery models.Recovery
}

// Insert creates a new Recovery row from a parsed YAML.
// Sets ImportedAt to now(), computes ExpiresAt = ImportedAt + 1y if YAML
// doesn't provide an explicit expires_at, and sets status to ACTIVE.
func Insert(db *gorm.DB, y *RecoveryYAML) (*models.Recovery, error) {
	now := time.Now().UTC()
	expiresAt := now.AddDate(1, 0, 0) // default: +1 year
	if y.ExpiresAt != "" {
		parsed, err := time.Parse(time.RFC3339, y.ExpiresAt)
		if err == nil {
			expiresAt = parsed.UTC()
		}
		// If parse fails, fall back to default — non-fatal
	}

	r := &models.Recovery{
		YamlID:           y.YamlID,
		YamlHash:         y.Hash,
		Name:             y.Name,
		Description:      y.Description,
		Shell:            y.Shell,
		TargetOS:         y.Target.OS,
		TargetVersionLte: y.Target.VersionLte,
		SuccessPattern:   y.SuccessPattern,
		MaxConcurrency:   y.MaxConcurrency,
		MaxAttempts:      y.MaxAttempts,
		RetryAfterSecs:   y.RetryAfter,
		AckTimeoutSecs:   y.AckTimeout,
		ExpiresAt:        expiresAt,
		DryRun:           y.DryRun,
		Script:           y.Script,
		DependsOnYamlID:  y.DependsOn,
		Status:           string(RecoveryStatusActive),
		ImportedAt:       now,
	}

	if err := db.Create(r).Error; err != nil {
		return nil, fmt.Errorf("insert recovery %q: %w", y.YamlID, err)
	}
	return r, nil
}

// FindByYAMLID returns the Recovery for a yaml_id, or (nil, gorm.ErrRecordNotFound).
func FindByYAMLID(db *gorm.DB, yamlID string) (*models.Recovery, error) {
	var r models.Recovery
	if err := db.Where("yaml_id = ?", yamlID).First(&r).Error; err != nil {
		return nil, err
	}
	return &r, nil
}

// AllRecoveries returns ALL recoveries regardless of status. Used at boot for
// FS reconciliation (we need to see ARCHIVED/BLOCKED ones too).
func AllRecoveries(db *gorm.DB) ([]models.Recovery, error) {
	var rs []models.Recovery
	if err := db.Find(&rs).Error; err != nil {
		return nil, fmt.Errorf("all recoveries: %w", err)
	}
	return rs, nil
}

// AllByStatus returns recoveries matching ANY of the given statuses.
func AllByStatus(db *gorm.DB, statuses ...RecoveryStatus) ([]models.Recovery, error) {
	strs := make([]string, len(statuses))
	for i, s := range statuses {
		strs[i] = string(s)
	}
	var rs []models.Recovery
	if err := db.Where("status IN ?", strs).Find(&rs).Error; err != nil {
		return nil, fmt.Errorf("recoveries by status: %w", err)
	}
	return rs, nil
}

// SetStatus updates the status (and optionally blocked_reason) of a recovery.
// Pass empty string for reason to clear it (e.g. when transitioning OUT of BLOCKED).
func SetStatus(db *gorm.DB, id uint, status RecoveryStatus, blockedReason string) error {
	updates := map[string]interface{}{
		"status":         string(status),
		"blocked_reason": blockedReason,
	}
	if err := db.Model(&models.Recovery{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("set recovery %d status %q: %w", id, status, err)
	}
	return nil
}

// UpdateContent updates script-related fields when re-importing.
// In v1 this refreshes ImportedAt on hash-match paths.
func UpdateContent(db *gorm.DB, id uint, y *RecoveryYAML) error {
	updates := map[string]interface{}{
		"yaml_hash":          y.Hash,
		"name":               y.Name,
		"description":        y.Description,
		"shell":              y.Shell,
		"target_os":          y.Target.OS,
		"target_version_lte": y.Target.VersionLte,
		"success_pattern":    y.SuccessPattern,
		"max_concurrency":    y.MaxConcurrency,
		"max_attempts":       y.MaxAttempts,
		"retry_after_secs":   y.RetryAfter,
		"ack_timeout_secs":   y.AckTimeout,
		"dry_run":            y.DryRun,
		"script":             y.Script,
		"depends_on_yaml_id": y.DependsOn,
		"imported_at":        time.Now().UTC(),
	}
	if err := db.Model(&models.Recovery{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("update recovery %d content: %w", id, err)
	}
	return nil
}

// DeleteRecovery hard-deletes a recovery and its targets in a transaction.
// Returns count of deleted target rows.
func DeleteRecovery(db *gorm.DB, id uint) (int64, error) {
	var targetCount int64
	tx := db.Begin()
	if tx.Error != nil {
		return 0, fmt.Errorf("begin tx for delete recovery %d: %w", id, tx.Error)
	}

	// Count targets first for audit logging.
	if err := tx.Model(&models.RecoveryTarget{}).Where("recovery_id = ?", id).Count(&targetCount).Error; err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("count targets for recovery %d: %w", id, err)
	}

	// Hard-delete targets.
	if err := tx.Unscoped().Where("recovery_id = ?", id).Delete(&models.RecoveryTarget{}).Error; err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("delete targets for recovery %d: %w", id, err)
	}

	// Hard-delete recovery.
	if err := tx.Unscoped().Where("id = ?", id).Delete(&models.Recovery{}).Error; err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("delete recovery %d: %w", id, err)
	}

	if err := tx.Commit().Error; err != nil {
		return 0, fmt.Errorf("commit delete recovery %d: %w", id, err)
	}
	return targetCount, nil
}

// AllAgents returns every non-deleted agent. Snapshot used at boot for
// initial target resolution.
func AllAgents(db *gorm.DB) ([]models.Agent, error) {
	var agents []models.Agent
	if err := db.Where("deleted_at IS NULL").Find(&agents).Error; err != nil {
		return nil, fmt.Errorf("all agents: %w", err)
	}
	return agents, nil
}

// FindAgentByID returns the agent (or err if not found / deleted).
func FindAgentByID(db *gorm.DB, agentID uint) (*models.Agent, error) {
	var a models.Agent
	if err := db.Where("id = ? AND deleted_at IS NULL", agentID).First(&a).Error; err != nil {
		return nil, fmt.Errorf("find agent %d: %w", agentID, err)
	}
	return &a, nil
}

// InsertMissingTargets uses ON CONFLICT DO NOTHING to create missing
// recovery_targets rows. Existing targets are NEVER touched (preserves
// terminal states per D7).
func InsertMissingTargets(db *gorm.DB, recoveryID uint, agentIDs []uint) error {
	if len(agentIDs) == 0 {
		return nil
	}

	rows := make([]models.RecoveryTarget, 0, len(agentIDs))
	for _, aid := range agentIDs {
		rows = append(rows, models.RecoveryTarget{
			RecoveryID: recoveryID,
			AgentID:    aid,
			Status:     string(TargetStatusPending),
			Attempts:   0,
		})
	}

	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error; err != nil {
		return fmt.Errorf("insert missing targets for recovery %d: %w", recoveryID, err)
	}
	return nil
}

// PendingForAgent returns all PENDING targets belonging to ACTIVE recoveries
// for the given agent, eagerly loading the parent Recovery.
// Sorted by recovery_id for deterministic ordering.
//
// Uses Joins for WHERE filtering (no column collision — only recovery_targets.*
// is selected) and Preload to fetch parent recoveries in a separate query.
// This avoids the GORM column-collision bug where SELECT rt.*, r.* causes the
// second occurrence of id/created_at/updated_at/deleted_at to overwrite the
// first, corrupting target.ID with recovery.ID and breaking all downstream
// DB updates.
func PendingForAgent(db *gorm.DB, agentID uint) ([]TargetWithRecovery, error) {
	var targets []models.RecoveryTarget
	err := db.
		Joins("JOIN recoveries r ON r.id = recovery_targets.recovery_id AND r.deleted_at IS NULL AND r.status = ?", string(RecoveryStatusActive)).
		Where("recovery_targets.agent_id = ? AND recovery_targets.status = ? AND recovery_targets.deleted_at IS NULL",
			agentID, string(TargetStatusPending)).
		Order("recovery_targets.recovery_id ASC").
		Preload("Recovery").
		Find(&targets).Error
	if err != nil {
		return nil, fmt.Errorf("pending targets for agent %d: %w", agentID, err)
	}

	result := make([]TargetWithRecovery, len(targets))
	for i, t := range targets {
		result[i] = TargetWithRecovery{Target: t, Recovery: t.Recovery}
	}
	return result, nil
}

// FindTarget returns one (recovery_id, agent_id) target.
func FindTarget(db *gorm.DB, recoveryID, agentID uint) (*models.RecoveryTarget, error) {
	var t models.RecoveryTarget
	if err := db.Where("recovery_id = ? AND agent_id = ?", recoveryID, agentID).First(&t).Error; err != nil {
		return nil, fmt.Errorf("find target recovery=%d agent=%d: %w", recoveryID, agentID, err)
	}
	return &t, nil
}

// FindTargetByCmdID returns the target whose last_cmd_id matches cmdID.
func FindTargetByCmdID(db *gorm.DB, cmdID string) (*models.RecoveryTarget, error) {
	var t models.RecoveryTarget
	if err := db.Where("last_cmd_id = ?", cmdID).First(&t).Error; err != nil {
		return nil, fmt.Errorf("find target by cmd_id %q: %w", cmdID, err)
	}
	return &t, nil
}

// SetTargetDispatched marks PENDING→DISPATCHED, increments attempts, sets
// last_attempt_at=now, and stores the cmd_id used for ACK matching. The
// UPDATE is guarded by status=PENDING so only one worker can claim a target.
// Returns rowsAffected so the caller can detect contention (rowsAffected=0
// means another worker took it; the caller MUST NOT proceed to send).
func SetTargetDispatched(db *gorm.DB, targetID uint, cmdID string) (int64, error) {
	now := time.Now().UTC()
	result := db.Model(&models.RecoveryTarget{}).
		Where("id = ? AND status = ?", targetID, string(TargetStatusPending)).
		Updates(map[string]interface{}{
			"status":          string(TargetStatusDispatched),
			"attempts":        gorm.Expr("attempts + 1"),
			"last_attempt_at": now,
			"last_cmd_id":     cmdID,
		})
	if result.Error != nil {
		return 0, fmt.Errorf("set target %d dispatched: %w", targetID, result.Error)
	}
	return result.RowsAffected, nil
}

// SetTargetSucceeded sets SUCCEEDED, completed_at=now, last_result=truncated output.
func SetTargetSucceeded(db *gorm.DB, targetID uint, result string) error {
	const maxResult = 8 * 1024 // 8 KB
	if len(result) > maxResult {
		result = result[:maxResult]
	}
	now := time.Now().UTC()
	err := db.Model(&models.RecoveryTarget{}).
		Where("id = ?", targetID).
		Updates(map[string]interface{}{
			"status":       string(TargetStatusSucceeded),
			"last_result":  result,
			"completed_at": now,
		}).Error
	if err != nil {
		return fmt.Errorf("set target %d succeeded: %w", targetID, err)
	}
	return nil
}

// SetTargetFailed sets FAILED, last_error=msg. If terminal=true, the target
// is at the FAILED terminal state (max_attempts reached).
func SetTargetFailed(db *gorm.DB, targetID uint, errMsg string, terminal bool) error {
	const maxErrLen = 512
	if len(errMsg) > maxErrLen {
		errMsg = errMsg[:maxErrLen]
	}
	updates := map[string]interface{}{
		"status":     string(TargetStatusFailed),
		"last_error": errMsg,
	}
	if terminal {
		now := time.Now().UTC()
		updates["completed_at"] = now
	}
	if err := db.Model(&models.RecoveryTarget{}).Where("id = ?", targetID).Updates(updates).Error; err != nil {
		return fmt.Errorf("set target %d failed: %w", targetID, err)
	}
	return nil
}

// SetTargetStatus sets an arbitrary status. Used by EXPIRED, SKIPPED, BLOCKED,
// and reset-to-PENDING (CLI retry).
func SetTargetStatus(db *gorm.DB, targetID uint, status TargetStatus) error {
	updates := map[string]interface{}{
		"status": string(status),
	}
	if status == TargetStatusSucceeded || status == TargetStatusFailed ||
		status == TargetStatusExpired || status == TargetStatusSkipped {
		now := time.Now().UTC()
		updates["completed_at"] = now
	}
	if err := db.Model(&models.RecoveryTarget{}).Where("id = ?", targetID).Updates(updates).Error; err != nil {
		return fmt.Errorf("set target %d status %q: %w", targetID, status, err)
	}
	return nil
}

// ResetTarget clears ALL completion-related state and sets the target back
// to PENDING with attempts=0. Used by CLI retry. Uses Select() to force
// nullable fields (last_attempt_at, completed_at) to be written as NULL —
// otherwise GORM's Updates(map) drops nil values and the row would keep
// stale terminal-state timestamps.
func ResetTarget(db *gorm.DB, targetID uint) error {
	err := db.Model(&models.RecoveryTarget{}).
		Where("id = ?", targetID).
		Select("attempts", "status", "last_error", "last_cmd_id",
			"last_attempt_at", "last_result", "completed_at").
		Updates(map[string]interface{}{
			"attempts":        0,
			"status":          string(TargetStatusPending),
			"last_error":      "",
			"last_cmd_id":     "",
			"last_attempt_at": nil,
			"last_result":     "",
			"completed_at":    nil,
		}).Error
	if err != nil {
		return fmt.Errorf("reset target %d: %w", targetID, err)
	}
	return nil
}

// CountTargetsByStatus returns counts grouped by status for a recovery.
func CountTargetsByStatus(db *gorm.DB, recoveryID uint) (map[TargetStatus]int, error) {
	type statusCount struct {
		Status string
		Count  int
	}
	var rows []statusCount
	err := db.Model(&models.RecoveryTarget{}).
		Select("status, COUNT(*) as count").
		Where("recovery_id = ?", recoveryID).
		Group("status").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("count targets by status for recovery %d: %w", recoveryID, err)
	}
	result := make(map[TargetStatus]int, len(rows))
	for _, r := range rows {
		result[TargetStatus(r.Status)] = r.Count
	}
	return result, nil
}

// ListTargets returns all targets for a recovery, sorted by agent_id.
func ListTargets(db *gorm.DB, recoveryID uint) ([]models.RecoveryTarget, error) {
	var ts []models.RecoveryTarget
	if err := db.Where("recovery_id = ?", recoveryID).Order("agent_id ASC").Find(&ts).Error; err != nil {
		return nil, fmt.Errorf("list targets for recovery %d: %w", recoveryID, err)
	}
	return ts, nil
}

// DispatchedPastTimeout returns DISPATCHED targets whose last_attempt_at is
// older than now - ack_timeout. Joins recoveries to apply per-row timeout.
//
// Uses Joins for WHERE filtering (no column collision — only recovery_targets.*
// is selected) and Preload to fetch parent recoveries in a separate query.
// Postgres-specific: uses interval arithmetic. COALESCE guards against NULL
// ack_timeout_secs (defaults to 300s) so a NULL doesn't silently stall the
// safety net (NULL in the predicate would evaluate UNKNOWN → no rows).
func DispatchedPastTimeout(db *gorm.DB, now time.Time) ([]TargetWithRecovery, error) {
	var targets []models.RecoveryTarget
	err := db.
		Joins("JOIN recoveries r ON r.id = recovery_targets.recovery_id AND r.deleted_at IS NULL").
		Where("recovery_targets.status = ? AND recovery_targets.deleted_at IS NULL", string(TargetStatusDispatched)).
		Where("recovery_targets.last_attempt_at IS NOT NULL").
		Where("recovery_targets.last_attempt_at < (?::timestamptz) - (COALESCE(r.ack_timeout_secs, 300) * interval '1 second')", now).
		Preload("Recovery").
		Find(&targets).Error
	if err != nil {
		return nil, fmt.Errorf("dispatched past timeout: %w", err)
	}

	result := make([]TargetWithRecovery, len(targets))
	for i, t := range targets {
		result[i] = TargetWithRecovery{Target: t, Recovery: t.Recovery}
	}
	return result, nil
}

// EligibleForRetry returns FAILED-but-not-terminal targets whose
// last_attempt_at + retry_after has elapsed AND attempts < max_attempts.
//
// Uses Joins for WHERE filtering (no column collision — only recovery_targets.*
// is selected) and Preload to fetch parent recoveries in a separate query.
// This was the primary failure path in E2E TEST 9: the SQL returned rows
// correctly, but Go received wrong target.ID values due to column collision,
// making SetTargetStatus update the wrong row (or no row at all).
// Postgres-specific: interval arithmetic with COALESCE(retry_after_secs, 1800)
// to avoid NULL → UNKNOWN predicate stalling retry detection.
func EligibleForRetry(db *gorm.DB, now time.Time) ([]TargetWithRecovery, error) {
	var targets []models.RecoveryTarget
	err := db.
		Joins("JOIN recoveries r ON r.id = recovery_targets.recovery_id AND r.deleted_at IS NULL").
		Where("recovery_targets.status = ? AND recovery_targets.deleted_at IS NULL", string(TargetStatusFailed)).
		Where("recovery_targets.attempts < r.max_attempts").
		Where("recovery_targets.last_attempt_at IS NOT NULL").
		Where("recovery_targets.last_attempt_at < (?::timestamptz) - (COALESCE(r.retry_after_secs, 1800) * interval '1 second')", now).
		Preload("Recovery").
		Find(&targets).Error
	if err != nil {
		return nil, fmt.Errorf("eligible for retry: %w", err)
	}

	result := make([]TargetWithRecovery, len(targets))
	for i, t := range targets {
		result[i] = TargetWithRecovery{Target: t, Recovery: t.Recovery}
	}
	return result, nil
}

// RollbackDispatchedToPending sets all DISPATCHED targets to PENDING and
// decrements attempts (shutdown should not burn a retry attempt).
// Returns count of rolled-back rows.
func RollbackDispatchedToPending(db *gorm.DB) (int64, error) {
	result := db.Model(&models.RecoveryTarget{}).
		Where("status = ?", string(TargetStatusDispatched)).
		Updates(map[string]interface{}{
			"status":   string(TargetStatusPending),
			"attempts": gorm.Expr("GREATEST(0, attempts - 1)"),
		})
	if result.Error != nil {
		return 0, fmt.Errorf("rollback dispatched to pending: %w", result.Error)
	}
	return result.RowsAffected, nil
}

// AboutToExpire returns ACTIVE or BLOCKED recoveries whose expires_at < now.
// Called by runSafetyNetTick BEFORE MarkExpiredRecoveries so that yaml_ids are
// known for logging the recovery_expired event per spec Observability requirement.
func AboutToExpire(db *gorm.DB, now time.Time) ([]models.Recovery, error) {
	var rs []models.Recovery
	err := db.Where("status IN ? AND expires_at < ?",
		[]string{string(RecoveryStatusActive), string(RecoveryStatusBlocked)}, now).
		Find(&rs).Error
	if err != nil {
		return nil, fmt.Errorf("about to expire: %w", err)
	}
	return rs, nil
}

// MarkExpiredRecoveries scans recoveries with status ACTIVE/BLOCKED whose
// expires_at < now and sets their status to EXPIRED. Returns count.
func MarkExpiredRecoveries(db *gorm.DB, now time.Time) (int64, error) {
	result := db.Model(&models.Recovery{}).
		Where("status IN ? AND expires_at < ?",
			[]string{string(RecoveryStatusActive), string(RecoveryStatusBlocked)}, now).
		Updates(map[string]interface{}{
			"status": string(RecoveryStatusExpired),
		})
	if result.Error != nil {
		return 0, fmt.Errorf("mark expired recoveries: %w", result.Error)
	}
	return result.RowsAffected, nil
}

// ExpireRecoveriesAndTargets atomically marks ACTIVE/BLOCKED recoveries with
// expires_at < now as EXPIRED and bulk-expires their non-terminal targets, all
// inside a single transaction. Returns the yaml_ids of recoveries that
// actually transitioned (i.e. were ACTIVE or BLOCKED at the start of this
// call), for per-recovery logging by the caller.
//
// This replaces the previous 3-query non-atomic flow (AboutToExpire +
// MarkExpiredRecoveries + ExpireTargetsOfRecoveries) which had TOCTOU races:
// rows could change between the SELECT and UPDATE, and newly-eligible
// recoveries could be missed by the targets sweep.
//
// Atomicity strategy: SELECT FOR UPDATE locks the candidate rows (only those
// currently in ACTIVE/BLOCKED state), captures their IDs/yaml_ids, then the
// UPDATE targets ONLY those locked IDs. This guarantees the returned set
// contains ONLY rows that genuinely transitioned in this call, never
// previously-EXPIRED rows whose expires_at is permanently in the past.
func ExpireRecoveriesAndTargets(db *gorm.DB, now time.Time) ([]string, error) {
	var transitionedYamlIDs []string
	err := db.Transaction(func(tx *gorm.DB) error {
		// 1. Lock and capture candidates: only ACTIVE/BLOCKED past expires_at.
		//    Rows already EXPIRED are excluded by the WHERE clause, so they
		//    cannot leak into the returned yaml_id list on subsequent ticks.
		var candidates []models.Recovery
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("status IN ? AND expires_at < ?",
				[]string{string(RecoveryStatusActive), string(RecoveryStatusBlocked)}, now).
			Find(&candidates).Error; err != nil {
			return fmt.Errorf("select candidates for expiration: %w", err)
		}
		if len(candidates) == 0 {
			return nil
		}

		ids := make([]uint, len(candidates))
		transitionedYamlIDs = make([]string, len(candidates))
		for i, r := range candidates {
			ids[i] = r.ID
			transitionedYamlIDs[i] = r.YamlID
		}

		// 2. Bulk update those exact IDs to EXPIRED.
		if err := tx.Model(&models.Recovery{}).
			Where("id IN ?", ids).
			Updates(map[string]interface{}{"status": string(RecoveryStatusExpired)}).Error; err != nil {
			return fmt.Errorf("update recoveries to expired: %w", err)
		}

		// 3. Bulk-expire non-terminal targets of those exact recoveries.
		terminalStates := []string{
			string(TargetStatusSucceeded),
			string(TargetStatusFailed),
			string(TargetStatusExpired),
			string(TargetStatusSkipped),
			string(TargetStatusBlocked),
		}
		return tx.Model(&models.RecoveryTarget{}).
			Where("recovery_id IN ? AND status NOT IN ?", ids, terminalStates).
			Updates(map[string]interface{}{
				"status":       string(TargetStatusExpired),
				"completed_at": now,
			}).Error
	})
	if err != nil {
		return nil, fmt.Errorf("expire recoveries and targets: %w", err)
	}
	return transitionedYamlIDs, nil
}

// ExpireTargetsOfRecoveries bulk-updates recovery_targets rows to EXPIRED for
// non-terminal targets belonging to the given recovery IDs. Sets completed_at
// to now so the row's lifecycle has a clean closing timestamp.
// Terminal states (SUCCEEDED, FAILED, EXPIRED, SKIPPED, BLOCKED) are preserved.
// Per spec R8: "Non-terminal targets of an expired recovery MUST transition to EXPIRED."
//
// DEPRECATED: superseded by ExpireRecoveriesAndTargets which performs both
// the recovery transition and the target sweep atomically. Retained for
// compatibility; not called from runSafetyNetTick anymore.
func ExpireTargetsOfRecoveries(db *gorm.DB, recoveryIDs []uint, now time.Time) (int64, error) {
	if len(recoveryIDs) == 0 {
		return 0, nil
	}
	terminalStates := []string{
		string(TargetStatusSucceeded),
		string(TargetStatusFailed),
		string(TargetStatusExpired),
		string(TargetStatusSkipped),
		string(TargetStatusBlocked),
	}
	result := db.Model(&models.RecoveryTarget{}).
		Where("recovery_id IN ? AND status NOT IN ?", recoveryIDs, terminalStates).
		Updates(map[string]interface{}{
			"status":       string(TargetStatusExpired),
			"completed_at": now,
		})
	if result.Error != nil {
		return 0, fmt.Errorf("expire targets of recoveries: %w", result.Error)
	}
	return result.RowsAffected, nil
}
