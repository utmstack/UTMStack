package recovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/utmstack/UTMStack/agent-manager/models"
	"gorm.io/gorm"
)

func BootReconcile(ctx context.Context, db *gorm.DB) error {
	// Step 1+2: Read and parse YAMLs from FS.
	dir := recoveryDir()
	rawFiles, readErrs := ReadDir(dir)
	for _, pe := range readErrs {
		LogRecoverySkippedInvalid(pe.Path, "read_error", pe.Err)
	}

	parsed, parseErrs := ParseAll(rawFiles)
	for _, pe := range parseErrs {
		LogRecoverySkippedInvalid(pe.Path, "parse_error", pe.Err)
	}

	// Step 3: Validate schema + sentinel for each parsed YAML.
	valid := make([]RecoveryYAML, 0, len(parsed))
	for i := range parsed {
		y := &parsed[i]
		if err := ValidateSchema(y); err != nil {
			LogRecoverySkippedInvalid(y.YamlID, "schema_invalid", err)
			continue
		}
		if err := ValidateSentinel(y); err != nil {
			LogRecoverySkippedInvalid(y.YamlID, "sentinel_invalid", err)
			continue
		}
		valid = append(valid, *y)
	}

	// Build a set of valid yaml_ids and a set of all parsed yaml_ids (for step 7).
	validByID := make(map[string]*RecoveryYAML, len(valid))
	for i := range valid {
		validByID[valid[i].YamlID] = &valid[i]
	}
	parsedIDSet := make(map[string]struct{}, len(parsed))
	for _, p := range parsed {
		parsedIDSet[p.YamlID] = struct{}{}
	}

	// Step 4: Cycle detection on valid YAMLs only.
	cycles := DetectCycles(valid) // map[yamlID]cyclePath

	// Fetch all recoveries from DB for reconciliation.
	dbRecoveries, err := AllRecoveries(db)
	if err != nil {
		return fmt.Errorf("boot reconcile: fetch recoveries: %w", err)
	}
	dbByID := make(map[string]*models.Recovery, len(dbRecoveries))
	for i := range dbRecoveries {
		dbByID[dbRecoveries[i].YamlID] = &dbRecoveries[i]
	}

	// Step 5: Hash-reconcile.
	for _, y := range valid {
		existing, exists := dbByID[y.YamlID]
		if !exists {
			// New in FS, not in DB → INSERT.
			inserted, insertErr := Insert(db, &y)
			if insertErr != nil {
				return fmt.Errorf("boot reconcile: insert %q: %w", y.YamlID, insertErr)
			}
			LogRecoveryImported(y.YamlID, y.Hash, 0)
			// Add to dbByID so subsequent steps see it.
			dbByID[y.YamlID] = inserted
		} else if existing.YamlHash != y.Hash {
			// Same yaml_id, different hash → log drift, keep DB content (Q3).
			LogRecoveryHashDrift(y.YamlID, existing.YamlHash, y.Hash)
			// No DB change — old content preserved.
		}
		// Same hash → no-op.
	}

	// Re-fetch dbByID after inserts to ensure fresh data.
	dbRecoveries, err = AllRecoveries(db)
	if err != nil {
		return fmt.Errorf("boot reconcile: re-fetch recoveries: %w", err)
	}
	dbByID = make(map[string]*models.Recovery, len(dbRecoveries))
	for i := range dbRecoveries {
		dbByID[dbRecoveries[i].YamlID] = &dbRecoveries[i]
	}

	// Step 6: Apply BLOCKED / unblock (self-healing) for non-ARCHIVED recoveries.
	for _, rec := range dbRecoveries {
		if RecoveryStatus(rec.Status) == RecoveryStatusArchived {
			continue // handled in step 8
		}

		cyclePath, inCycle := cycles[rec.YamlID]
		depMissing := false
		if rec.DependsOnYamlID != "" {
			_, depInValid := validByID[rec.DependsOnYamlID]
			if !depInValid {
				depMissing = true
			}
		}

		if inCycle {
			reason := BlockedReasonCycleDetected + ":" + cyclePath
			if RecoveryStatus(rec.Status) != RecoveryStatusBlocked || rec.BlockedReason != reason {
				if setErr := SetStatus(db, rec.ID, RecoveryStatusBlocked, reason); setErr != nil {
					return fmt.Errorf("boot reconcile: block cycle %q: %w", rec.YamlID, setErr)
				}
				parts := strings.SplitN(reason, ":", 2)
				detail := ""
				if len(parts) == 2 {
					detail = parts[1]
				}
				LogRecoveryBlocked(rec.YamlID, BlockedReasonCycleDetected, detail)
			}
		} else if depMissing {
			reason := BlockedReasonDependencyMissing + ":" + rec.DependsOnYamlID
			if RecoveryStatus(rec.Status) != RecoveryStatusBlocked || rec.BlockedReason != reason {
				if setErr := SetStatus(db, rec.ID, RecoveryStatusBlocked, reason); setErr != nil {
					return fmt.Errorf("boot reconcile: block dep-missing %q: %w", rec.YamlID, setErr)
				}
				LogRecoveryBlocked(rec.YamlID, BlockedReasonDependencyMissing, rec.DependsOnYamlID)
			}
		} else if RecoveryStatus(rec.Status) == RecoveryStatusBlocked {
			// Self-heal: was BLOCKED but now conditions are resolved.
			if setErr := SetStatus(db, rec.ID, RecoveryStatusActive, ""); setErr != nil {
				return fmt.Errorf("boot reconcile: unblock %q: %w", rec.YamlID, setErr)
			}
			LogRecoveryUnblocked(rec.YamlID)
			LogRecoveryReactivated(rec.YamlID, false)
		}
	}

	// Step 7: Archive removed — yaml_id in DB (non-ARCHIVED, non-EXPIRED) but not in FS.
	for _, rec := range dbRecoveries {
		status := RecoveryStatus(rec.Status)
		if status == RecoveryStatusArchived || status == RecoveryStatusExpired {
			continue
		}
		// "truly absent" = not in parsed set AND not in valid set.
		_, inParsed := parsedIDSet[rec.YamlID]
		_, inValid := validByID[rec.YamlID]
		if !inParsed && !inValid {
			if setErr := SetStatus(db, rec.ID, RecoveryStatusArchived, ""); setErr != nil {
				return fmt.Errorf("boot reconcile: archive %q: %w", rec.YamlID, setErr)
			}
			LogRecoveryArchived(rec.YamlID)
		}
	}

	// Step 8: Reactivate returning — yaml_id was ARCHIVED but is back in FS.
	for _, rec := range dbRecoveries {
		if RecoveryStatus(rec.Status) != RecoveryStatusArchived {
			continue
		}
		y, inValid := validByID[rec.YamlID]
		if !inValid {
			continue
		}
		// It's back — reactivate.
		hashDrift := rec.YamlHash != y.Hash
		if setErr := SetStatus(db, rec.ID, RecoveryStatusActive, ""); setErr != nil {
			return fmt.Errorf("boot reconcile: reactivate %q: %w", rec.YamlID, setErr)
		}
		if hashDrift {
			LogRecoveryHashDrift(rec.YamlID, rec.YamlHash, y.Hash)
		}
		LogRecoveryReactivated(rec.YamlID, hashDrift)
	}

	// Step 9: Resolve targets for each ACTIVE recovery.
	// Fetch the freshest set of recoveries after all status changes above.
	activeRecoveries, err := AllByStatus(db, RecoveryStatusActive)
	if err != nil {
		return fmt.Errorf("boot reconcile: fetch active recoveries: %w", err)
	}
	agents, err := AllAgents(db)
	if err != nil {
		return fmt.Errorf("boot reconcile: fetch agents: %w", err)
	}

	for _, r := range activeRecoveries {
		if resolveErr := resolveTargetsForRecovery(db, &r, agents); resolveErr != nil {
			return fmt.Errorf("boot reconcile: resolve targets for %q: %w", r.YamlID, resolveErr)
		}
	}

	return nil
}

// ResolveTargetsForAgent filters all ACTIVE recoveries against a single agent
// and inserts missing targets. Called by hooks when an agent registers or updates.
func ResolveTargetsForAgent(db *gorm.DB, agent *models.Agent) error {
	if agent == nil || agent.DeletedAt != nil {
		return nil
	}
	activeRecoveries, err := AllByStatus(db, RecoveryStatusActive)
	if err != nil {
		return fmt.Errorf("resolve targets for agent %d: %w", agent.ID, err)
	}

	for _, r := range activeRecoveries {
		matched := Resolve(&r, []models.Agent{*agent})
		if len(matched) == 0 {
			continue
		}
		ids := []uint{agent.ID}
		if insErr := InsertMissingTargets(db, r.ID, ids); insErr != nil {
			return fmt.Errorf("insert target for recovery %q agent %d: %w", r.YamlID, agent.ID, insErr)
		}
	}
	return nil
}

// resolveTargetsForRecovery resolves targets for one active recovery against
// all agents, then bulk-inserts missing targets. Used by step 9 of BootReconcile.
func resolveTargetsForRecovery(db *gorm.DB, r *models.Recovery, agents []models.Agent) error {
	matched := Resolve(r, agents)
	if len(matched) == 0 {
		return nil
	}
	ids := make([]uint, 0, len(matched))
	for _, a := range matched {
		ids = append(ids, a.ID)
	}
	return InsertMissingTargets(db, r.ID, ids)
}

// recoveryDir returns the configured recovery directory (from config package).
// Separated into a small helper to keep BootReconcile import-clean.
func recoveryDir() string {
	// Avoids importing config in every caller; config is imported via this package's
	// recovery.go entrypoint which wires the dir at Init time if needed.
	// For now, delegate directly to config.
	return configRecoveryDir
}
