package usecase

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"gorm.io/gorm"
)

type FlowBootstrap struct {
	srcDir string // image source flows (e.g. /utmstack/soar)
	store  *FlowStore
	db     *gorm.DB
}

func NewFlowBootstrap(srcDir string, store *FlowStore, db *gorm.DB) *FlowBootstrap {
	return &FlowBootstrap{srcDir: srcDir, store: store, db: db}
}

func (b *FlowBootstrap) Run(ctx context.Context) error {

	if err := b.purgeLooseFlows(); err != nil {
		_ = catcher.Error("soar: purging loose flows failed", err, nil)
	}

	if err := b.seedSystemOverlay(); err != nil {
		_ = catcher.Error("soar: seeding system flows failed", err, nil)
	}

	if err := b.store.Load(); err != nil {
		return err
	}

	if b.db != nil {
		if err := b.migrateLegacyFlows(ctx); err != nil {
			_ = catcher.Error("soar: legacy flow migration failed", err, nil)
		}
		if err := b.store.Load(); err != nil {
			return err
		}
	}

	return nil
}

func (b *FlowBootstrap) purgeLooseFlows() error {
	root := filepath.Dir(b.store.systemDir)
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.Name() == SystemSubdir || e.Name() == UserSubdir {
			continue
		}
		if err := os.RemoveAll(filepath.Join(root, e.Name())); err != nil {
			_ = catcher.Error("soar: removing loose flow entry failed", err, map[string]any{"entry": e.Name()})
		}
	}
	return nil
}

func (b *FlowBootstrap) seedSystemOverlay() error {
	if _, err := os.Stat(b.srcDir); os.IsNotExist(err) {
		return nil
	}

	expected := make(map[string]bool)

	err := filepath.WalkDir(b.srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) != FlowFileExt {
			return nil
		}
		rel, err := filepath.Rel(b.srcDir, path)
		if err != nil {
			return nil
		}
		expected[rel] = true
		target := filepath.Join(b.store.systemDir, rel)

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return nil
		}

		// Preserve the operator's disabled state but refresh the content into
		// whichever file exists, so updates reach disabled flows too.
		if disabled := target + DisabledSuffix; fileExists(disabled) {
			_ = os.WriteFile(disabled, data, 0o644)
		} else {
			_ = os.WriteFile(target, data, 0o644)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return b.pruneSystemOverlay(expected)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (b *FlowBootstrap) pruneSystemOverlay(expected map[string]bool) error {
	if _, err := os.Stat(b.store.systemDir); os.IsNotExist(err) {
		return nil
	}
	return filepath.WalkDir(b.store.systemDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, relErr := filepath.Rel(b.store.systemDir, path)
		if relErr != nil {
			return nil
		}
		canon := strings.TrimSuffix(rel, DisabledSuffix)
		if expected[canon] {
			return nil
		}
		if rmErr := os.Remove(path); rmErr != nil {
			_ = catcher.Error("soar: pruning orphaned system flow failed", rmErr, map[string]any{"flow": rel})
		}
		return nil
	})
}

func (b *FlowBootstrap) migrateLegacyFlows(ctx context.Context) error {
	if !b.db.Migrator().HasTable(&domain.AlertResponseRule{}) {
		return nil
	}

	var legacy []domain.AlertResponseRule
	if err := b.db.WithContext(ctx).Preload("Templates").Find(&legacy).Error; err != nil {
		return err
	}
	if len(legacy) == 0 {
		return nil
	}

	systemByName := make(map[string]string)
	b.store.mu.RLock()
	for _, sf := range b.store.flows {
		if sf.system {
			systemByName[sf.Name] = sf.RelPath
		}
	}
	b.store.mu.RUnlock()

	failed := 0
	for i := range legacy {
		row := &legacy[i]

		if row.SystemOwner {
			if !row.RuleActive {
				if relPath, ok := systemByName[row.RuleName]; ok {
					if err := b.store.SetEnabled(relPath, false); err != nil {
						_ = catcher.Error("soar: reconciling disabled system flow failed", err, map[string]any{"flow": row.RuleName})
						failed++
					}
				}
			}
			continue
		}

		created, err := b.store.Create(legacyToFlow(row))
		if err != nil {
			if os.IsExist(err) {
				continue // already migrated
			}
			_ = catcher.Error("soar: migrating legacy user flow failed", err, map[string]any{"flow": row.RuleName})
			failed++
			continue
		}
		if !row.RuleActive {
			if err := b.store.SetEnabled(created.RelPath, false); err != nil {
				_ = catcher.Error("soar: disabling migrated user flow failed", err, map[string]any{"flow": row.RuleName})
				failed++
			}
		}
	}

	if failed > 0 {
		catcher.Warn("soar: legacy flow migration had failures — keeping utm_alert_response_rule for retry on next boot", nil)
		return nil
	}
	return b.dropLegacyFlowTables()
}

func (b *FlowBootstrap) dropLegacyFlowTables() error {
	if err := b.db.Exec("DROP TABLE IF EXISTS utm_alert_response_rule_template CASCADE").Error; err != nil {
		return err
	}
	if err := b.db.Exec("DROP TABLE IF EXISTS utm_alert_response_rule CASCADE").Error; err != nil {
		return err
	}
	catcher.Info("soar: legacy alert-response-rule table migrated to YAML and dropped", nil)
	return nil
}

func legacyToFlow(row *domain.AlertResponseRule) Flow {
	var conds []FlowCondition
	if strings.TrimSpace(row.RuleConditions) != "" {
		var ft []domain.FilterType
		if err := json.Unmarshal([]byte(row.RuleConditions), &ft); err == nil {
			for _, f := range ft {
				conds = append(conds, FlowCondition{Operator: string(f.Operator), Field: f.Field, Value: f.Value})
			}
		}
	}

	var excluded []string
	for _, a := range strings.Split(row.ExcludedAgents, ",") {
		if a = strings.TrimSpace(a); a != "" {
			excluded = append(excluded, a)
		}
	}

	var commands []FlowCommand
	seen := make(map[string]bool)
	always := domain.ConditionAlways
	addCmd := func(c string) {
		if c = strings.TrimSpace(c); c != "" && !seen[c] {
			seen[c] = true
			fc := FlowCommand{Command: c}
			if len(commands) > 0 {
				fc.Condition = &always
			}
			commands = append(commands, fc)
		}
	}
	addCmd(row.RuleCmd)
	for _, t := range row.Templates {
		addCmd(t.Command)
	}

	return Flow{
		Name:           row.RuleName,
		Description:    row.RuleDescription,
		Conditions:     conds,
		Commands:       commands,
		Shell:          row.RuleShell,
		AgentPlatform:  row.AgentPlatform,
		DefaultAgent:   row.DefaultAgent,
		ExcludedAgents: excluded,
	}
}
