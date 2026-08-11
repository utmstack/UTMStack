package repository

import (
	"os"
	"path/filepath"
	"sort"
	"sync"
	"syscall"
	"time"

	"context"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/domain"
	"gopkg.in/yaml.v3"
)

type engineConfig struct {
	dir string
	mu  sync.Mutex // in-process fast path; withFileLock handles cross-process safety
}

func NewEngineConfig(dir string) connectors.EngineConfigRepository {
	return &engineConfig{dir: dir}
}

func (w *engineConfig) readTenantFile() (domain.TenantsFile, error) {
	data, err := os.ReadFile(filepath.Join(w.dir, TenantFileName))
	if os.IsNotExist(err) {
		return domain.TenantsFile{}, nil
	}
	if err != nil {
		return domain.TenantsFile{}, err
	}
	var content domain.TenantsFile
	if err := yaml.Unmarshal(data, &content); err != nil {
		return domain.TenantsFile{}, err
	}
	return content, nil
}

func (w *engineConfig) readTenantEntry(tenantId string) (domain.TenantEntry, error) {
	content, err := w.readTenantFile()
	if err != nil {
		return domain.TenantEntry{}, err
	}
	if tenantId == "" {
		tenantId = DefaultTenantID
	}
	for _, t := range content.Tenants {
		if t.ID == tenantId {
			return t, nil
		}
	}
	// A tenant the file has not heard of yet starts empty rather than
	// inheriting whatever the first entry happens to say.
	name := ""
	if tenantId == DefaultTenantID {
		name = DefaultTenantName
	}
	return domain.TenantEntry{ID: tenantId, Name: name}, nil
}

// writeTenantsLocked refreshes the asset projection while leaving every
// tenant's own choices — what it disabled, the order it asked for — exactly as
// they were. Those are answers only that tenant can give.
func (w *engineConfig) writeTenantsLocked(tenants []domain.TenantAssets) error {
	content, err := w.readTenantFile()
	if err != nil {
		return err
	}
	prev := make(map[string]domain.TenantEntry, len(content.Tenants))
	for _, t := range content.Tenants {
		prev[t.ID] = t
	}

	out := make([]domain.TenantEntry, 0, len(tenants)+1)
	seenDefault := false
	for _, t := range tenants {
		if t.ID == "" {
			continue
		}
		if t.ID == DefaultTenantID {
			seenDefault = true
			if t.Name == "" {
				t.Name = DefaultTenantName
			}
		}
		e := prev[t.ID]
		e.ID, e.Name, e.Assets = t.ID, t.Name, t.Assets
		out = append(out, e)
	}
	if !seenDefault {
		e := prev[DefaultTenantID]
		e.ID, e.Name = DefaultTenantID, DefaultTenantName
		out = append(out, e)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })

	return w.writeYAML(TenantFileName, domain.TenantsFile{Tenants: out})
}

// writeTenantEntryLocked replaces one tenant's entry and leaves the others
// untouched, adding it if the file has never seen that tenant.
func (w *engineConfig) writeTenantEntryLocked(t domain.TenantEntry) error {
	content, err := w.readTenantFile()
	if err != nil {
		return err
	}
	out := make([]domain.TenantEntry, 0, len(content.Tenants)+1)
	replaced := false
	for _, e := range content.Tenants {
		if e.ID == t.ID {
			out = append(out, t)
			replaced = true
			continue
		}
		out = append(out, e)
	}
	if !replaced {
		out = append(out, t)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })

	return w.writeYAML(TenantFileName, domain.TenantsFile{Tenants: out})
}

func (w *engineConfig) WriteTenants(tenants []domain.TenantAssets) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return withFileLock(w.tenantLockPath(), func() error {
		return w.writeTenantsLocked(tenants)
	})
}

func (w *engineConfig) SetRuleDisabled(tenantId, identity string, disabled bool) error {
	return w.mutateTenant(tenantId, func(t *domain.TenantEntry) {
		t.DisabledRules = toggleIdentity(t.DisabledRules, identity, disabled)
	})
}

func (w *engineConfig) SetPipelineDisabled(tenantId, identity string, disabled bool) error {
	return w.mutateTenant(tenantId, func(t *domain.TenantEntry) {
		t.DisabledPipelines = toggleIdentity(t.DisabledPipelines, identity, disabled)
	})
}

// SetPipelineOrder records the whole sequence rather than one position: the
// engine reads it as a list of names, so a partial update would be ambiguous.
// An empty list clears the preference and the tenant goes back to the order the
// pipeline files declare.
func (w *engineConfig) SetPipelineOrder(tenantId string, order []string) error {
	return w.mutateTenant(tenantId, func(t *domain.TenantEntry) {
		t.PipelineOrder = order
	})
}

// mutateTenant applies a change to one tenant's entry under the file lock,
// leaving every other tenant's entry as it was.
func (w *engineConfig) mutateTenant(tenantId string, apply func(*domain.TenantEntry)) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return withFileLock(w.tenantLockPath(), func() error {
		t, err := w.readTenantEntry(tenantId)
		if err != nil {
			return err
		}
		apply(&t)
		return w.writeTenantEntryLocked(t)
	})
}

func (w *engineConfig) PokeRuleReload(absPath string) {
	now := time.Now()
	_ = os.Chtimes(absPath, now, now)
}

func (w *engineConfig) DisabledRuleSet(tenantId string) map[string]bool {
	return w.identitySet(tenantId, func(t domain.TenantEntry) []string { return t.DisabledRules })
}

func (w *engineConfig) DisabledPipelineSet(tenantId string) map[string]bool {
	return w.identitySet(tenantId, func(t domain.TenantEntry) []string { return t.DisabledPipelines })
}

func (w *engineConfig) PipelineOrder(tenantId string) []string {
	t, err := w.readTenantEntry(tenantId)
	if err != nil {
		return nil
	}
	return t.PipelineOrder
}

func (w *engineConfig) identitySet(tenantId string, pick func(domain.TenantEntry) []string) map[string]bool {
	out := make(map[string]bool)
	t, err := w.readTenantEntry(tenantId)
	if err != nil {
		return out
	}
	for _, id := range pick(t) {
		out[id] = true
	}
	return out
}

func toggleIdentity(set []string, identity string, add bool) []string {
	idx := -1
	for i, v := range set {
		if v == identity {
			idx = i
			break
		}
	}
	if add {
		if idx >= 0 {
			return set
		}
		return append(set, identity)
	}
	if idx < 0 {
		return set
	}
	return append(set[:idx], set[idx+1:]...)
}

func (w *engineConfig) tenantLockPath() string {
	return filepath.Join(w.dir, TenantFileName)
}

func (w *engineConfig) ReadPatterns() (map[string]string, error) {
	data, err := os.ReadFile(filepath.Join(w.dir, PatternsFileName))
	if os.IsNotExist(err) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, err
	}
	var content domain.PatternsFile
	if err := yaml.Unmarshal(data, &content); err != nil {
		return nil, err
	}
	if content.Patterns == nil {
		return map[string]string{}, nil
	}
	return content.Patterns, nil
}

func (w *engineConfig) WritePatterns(patterns map[string]string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if patterns == nil {
		patterns = map[string]string{}
	}
	return withFileLock(filepath.Join(w.dir, PatternsFileName), func() error {
		return w.writeYAML(PatternsFileName, domain.PatternsFile{Patterns: patterns})
	})
}

func (w *engineConfig) writeYAML(name string, v any) error {
	if err := os.MkdirAll(w.dir, 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	dst := filepath.Join(w.dir, name)
	tmp := dst + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}

func withFileLock(path string, fn func() error) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	lf, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	defer lf.Close()
	if err := syscall.Flock(int(lf.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	return fn()
}

type EngineConfigBootstrap struct {
	writer connectors.EngineConfigRepository
}

func NewEngineConfigBootstrap(writer connectors.EngineConfigRepository) *EngineConfigBootstrap {
	return &EngineConfigBootstrap{writer: writer}
}

func (b *EngineConfigBootstrap) Run(ctx context.Context) error {
	return b.writer.WritePatterns(b.writer.BuiltInPatterns())
}

var systemPatterns = map[string]string{
	"ciscoMacAddr":    `(?:(?:[A-Fa-f0-9]{4}\.){2}[A-Fa-f0-9]{4})`,
	"syslogDate":      `[A-Z][a-z]{2} \d{1,2} \d{2}:\d{2}:\d{2}`,
	"winMacAddr":      `(?:(?:[A-Fa-f0-9]{2}-){5}[A-Fa-f0-9]{2})`,
	"commonMacAddr":   `(?:(?:[A-Fa-f0-9]{2}:){5}[A-Fa-f0-9]{2})|(?:(?:[A-Fa-f0-9]{2}-){5}[A-Fa-f0-9]{2})`,
	"integer":         `(?:[+-]?(?:[0-9]+))`,
	"day":             `(?:Mon(?:day)?|Tue(?:sday)?|Wed(?:nesday)?|Thu(?:rsday)?|Fri(?:day)?|Sat(?:urday)?|Sun(?:day)?)`,
	"word":            `\b\w+\b`,
	"greedy":          `.*`,
	"space":           `\s+`,
	"notSpace":        `\S+`,
	"monthName":       `\b(?:[Jj]an(?:uary|uar)?|[Ff]eb(?:ruary|ruar)?|[Mm](?:a|ä)?r(?:ch|z)?|[Aa]pr(?:il)?|[Mm]a(?:y|i)?|[Jj]un(?:e|i)?|[Jj]ul(?:y|i)?|[Aa]ug(?:ust)?|[Ss]ep(?:tember)?|[Oo](?:c|k)?t(?:ober)?|[Nn]ov(?:ember)?|[Dd]e(?:c|z)(?:ember)?)\b`,
	"ipv4":            `(((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.)){3}((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)))`,
	"email":           `((?P<name>[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+)@(?P<domain>[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*))`,
	"domain":          `((?:[_a-z0-9](?:[_a-z0-9-]{0,61}[a-z0-9])?\.)+(?:[a-z](?:[a-z0-9-]{0,61}[a-z0-9])?)?)`,
	"hostname":        `(\b(?:[0-9A-Za-z][0-9A-Za-z-]{0,62})(?:\.(?:[0-9A-Za-z][0-9A-Za-z-]{0,62}))*(\.?|\b))`,
	"data":            `(.*?)`,
	"ipv6":            `([0-9a-fA-F]{1,4}(:[0-9a-fA-F]{0,4}){1,7}|::[0-1]?)`,
	"uuid":            `([A-Fa-f0-9]{8}-(?:[A-Fa-f0-9]{4}-){3}[A-Fa-f0-9]{12})`,
	"monthNumber":     `(?:0[1-9]|1[0-2])`,
	"monthDay":        `(?:(?:0[1-9])|(?:[12][0-9])|(?:3[01])|[1-9])`,
	"year":            `(([1-9])[0-9]{1,3})`,
	"hour":            `(([01][0-9])|2[0-4])`,
	"minute":          `(?:[0-5][0-9])`,
	"seconds":         `(?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)`,
	"time":            `((([01][0-9])|2[0-4]):(?:[0-5][0-9])(?::(?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)))`,
	"iso8601Timezone": `(Z|([+-](([01][0-9])|2[0-4]):?([0-5][0-9])))`,
}

func (w *engineConfig) BuiltInPatterns() map[string]string {
	out := make(map[string]string, len(systemPatterns))
	for id, def := range systemPatterns {
		out[id] = def
	}
	return out
}
