package usecase

import (
	"context"
	"encoding/json"
	tenant_domain "github.com/utmstack/utmstack/backend/modules/tenant/domain"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/utmstack/utmstack/backend/pkg/authz"

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
	if err := b.seedSystemOverlay(); err != nil {
		_ = catcher.Error("soar: seeding system flows failed", err, nil)
	}
	return b.store.Load()
}

// seedSystemOverlay refreshes every tenant's copy of the flows the image ships.
// Each tenant holding its own copy is what makes a release a plain overwrite:
// the content is replaced in whichever file exists, so a flow a tenant switched
// off stays off and still gets the update.
func (b *FlowBootstrap) seedSystemOverlay() error {
	tenants, err := b.tenants()
	if err != nil {
		return err
	}
	for _, t := range tenants {
		if err := b.seedSystemOverlayFor(t); err != nil {
			return err
		}
	}
	return nil
}

func (b *FlowBootstrap) tenants() ([]string, error) {
	var ids []string
	err := b.db.Model(&tenant_domain.Tenant{}).
		Where("status <> ?", "TERMINATED").
		Pluck("id", &ids).Error
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		// Before the first tenant row exists: the install has the one it was
		// born with.
		return []string{authz.DefaultTenantID}, nil
	}
	return ids, nil
}

func (b *FlowBootstrap) seedSystemOverlayFor(tenant string) error {
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
		target := filepath.Join(b.store.userDir, tenant, SystemSubdir, rel)

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

	return b.pruneSystemOverlay(tenant, expected)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (b *FlowBootstrap) pruneSystemOverlay(tenant string, expected map[string]bool) error {
	dir := filepath.Join(b.store.userDir, tenant, SystemSubdir)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, relErr := filepath.Rel(dir, path)
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
