package repository

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/domain"
	"gopkg.in/yaml.v3"
)

const reloadInterval = 10 * time.Second

type RuleStore struct {
	systemDir string
	userDir   string
	writer    connectors.EngineConfigRepository

	mu    sync.RWMutex
	rules []*domain.StoredRule          // load order (system first, then user)
	index map[string]*domain.StoredRule // relPath -> rule
}

func NewRuleStore(systemDir, userDir string, writer connectors.EngineConfigRepository) *RuleStore {
	return &RuleStore{
		systemDir: systemDir,
		userDir:   userDir,
		writer:    writer,
		index:     make(map[string]*domain.StoredRule),
	}
}

func ruleIdentity(relPath string) string {
	base := filepath.Base(relPath)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func (s *RuleStore) Load() error {
	// The shared cache carries the default view; each tenant's answer is
	// stamped when the rules are read, in List.
	disabled := s.writer.DisabledRuleSet(DefaultTenantID)

	rules := make([]*domain.StoredRule, 0, 256)
	index := make(map[string]*domain.StoredRule, 256)

	for _, ov := range []struct {
		dir    string
		system bool
	}{
		{s.systemDir, true},
		{s.userDir, false},
	} {
		loaded, err := loadOverlay(ov.dir, ov.system, disabled)
		if err != nil {
			return err
		}
		for _, sr := range loaded {
			if prev, ok := index[sr.RelPath]; ok {
				// User overlay overrides a system rule with the same relPath.
				*prev = *sr
				continue
			}
			index[sr.RelPath] = sr
			rules = append(rules, sr)
		}
	}

	s.mu.Lock()
	s.rules = rules
	s.index = index
	s.mu.Unlock()
	return nil
}

func (s *RuleStore) Watch(ctx context.Context) {
	ticker := time.NewTicker(reloadInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.Load(); err != nil {
				_ = catcher.Error("eventprocessing: periodic rule reload failed", err, nil)
			}
		}
	}
}

func loadOverlay(dir string, system bool, disabled map[string]bool) ([]*domain.StoredRule, error) {
	var out []*domain.StoredRule

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return out, nil
	}

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != RuleFileExt {
			return nil // ignore non-rule files
		}

		rule, rerr := readRuleFile(path)
		if rerr != nil {
			return nil // skip unparsable files rather than failing the whole load
		}

		rel, rerr := filepath.Rel(dir, path)
		if rerr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)

		var mod time.Time
		if info, ierr := d.Info(); ierr == nil {
			mod = info.ModTime()
		}

		out = append(out, &domain.StoredRule{
			Rule:     rule,
			RelPath:  rel,
			Modified: mod,
			System:   system,
			Enabled:  !disabled[ruleIdentity(rel)],
		})
		return nil
	})
	return out, err
}

func (s *RuleStore) FindByRelPath(relPath string) *domain.StoredRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if sr, ok := s.index[relPath]; ok {
		cp := *sr
		return &cp
	}
	return nil
}

func (s *RuleStore) FindByName(name string) *domain.StoredRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sr := range s.rules {
		if strings.EqualFold(sr.Name, name) {
			cp := *sr
			return &cp
		}
	}
	return nil
}

func (s *RuleStore) CorrelationsByName(name string) (json.RawMessage, bool) {
	sr := s.FindByName(name)
	if sr == nil {
		return nil, false
	}
	return anyToRaw(sr.Correlation), true
}

func (s *RuleStore) List(f connectors.RuleListFilter) ([]*domain.StoredRule, int) {
	disabled := s.writer.DisabledRuleSet(f.TenantId)

	s.mu.RLock()
	matched := make([]*domain.StoredRule, 0, len(s.rules))
	for _, sr := range s.rules {
		cp := *sr
		// Enabled is per tenant, so it is stamped here rather than trusted from
		// the shared cache — and stamped before the filter, so a query on the
		// active state answers for the tenant that asked.
		cp.Enabled = !disabled[ruleIdentity(cp.RelPath)]
		if ruleMatches(&cp, f) {
			matched = append(matched, &cp)
		}
	}
	s.mu.RUnlock()

	sort.SliceStable(matched, func(i, j int) bool {
		return strings.ToLower(matched[i].Name) < strings.ToLower(matched[j].Name)
	})

	total := len(matched)
	if f.Size <= 0 {
		return matched, total
	}
	start := f.Page * f.Size
	if start >= total {
		return []*domain.StoredRule{}, total
	}
	end := start + f.Size
	if end > total {
		end = total
	}
	return matched[start:end], total
}

func ruleMatches(sr *domain.StoredRule, f connectors.RuleListFilter) bool {
	if f.TenantId != "" && sr.TenantId != "" && sr.TenantId != f.TenantId {
		return false
	}
	if f.Name != "" && !strings.Contains(strings.ToLower(sr.Name), strings.ToLower(f.Name)) {
		return false
	}
	if f.Search != "" && !strings.Contains(strings.ToLower(sr.Name), strings.ToLower(f.Search)) {
		return false
	}
	if f.Active != nil && sr.Enabled != *f.Active {
		return false
	}
	if f.SystemOwner != nil && sr.System != *f.SystemOwner {
		return false
	}
	if len(f.Categories) > 0 && !containsStr(f.Categories, sr.Category) {
		return false
	}
	if len(f.Adversaries) > 0 && !containsAdversary(f.Adversaries, sr.Adversary) {
		return false
	}
	if len(f.Techniques) > 0 && !containsStr(f.Techniques, sr.Technique) {
		return false
	}
	if len(f.Confidentiality) > 0 && !containsInt(f.Confidentiality, sr.Impact.Confidentiality) {
		return false
	}
	if len(f.Integrity) > 0 && !containsInt(f.Integrity, sr.Impact.Integrity) {
		return false
	}
	if len(f.Availability) > 0 && !containsInt(f.Availability, sr.Impact.Availability) {
		return false
	}
	if len(f.DataTypes) > 0 && !anyStr(f.DataTypes, sr.DataTypes) {
		return false
	}
	if !f.InitDate.IsZero() && sr.Modified.Before(f.InitDate) {
		return false
	}
	if !f.EndDate.IsZero() && sr.Modified.After(f.EndDate) {
		return false
	}
	return true
}

func (s *RuleStore) withStoreLock(fn func() error) error {
	return withFileLock(filepath.Join(s.userDir, ".rules"), fn)
}

func (s *RuleStore) Create(rule domain.Rule, tenantId string) (*domain.StoredRule, error) {
	rule.TenantId = tenantId

	var relPath string
	if tenantId == "" {
		relPath = slug(rule.Name) + RuleFileExt
	} else {
		relPath = filepath.ToSlash(filepath.Join(tenantId, slug(rule.Name)+"-"+tenantId+RuleFileExt))
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	abs := filepath.Join(s.userDir, filepath.FromSlash(relPath))
	if err := s.withStoreLock(func() error {
		if _, err := os.Stat(abs); err == nil {
			return os.ErrExist
		}
		return writeRuleFile(abs, rule)
	}); err != nil {
		return nil, err
	}

	sr := &domain.StoredRule{
		Rule:     rule,
		RelPath:  relPath,
		Modified: fileModTime(abs),
		System:   false,
		Enabled:  true,
	}
	s.index[relPath] = sr
	s.rules = append(s.rules, sr)

	cp := *sr
	return &cp, nil
}

func (s *RuleStore) Update(relPath string, rule domain.Rule) (*domain.StoredRule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sr, ok := s.index[relPath]
	if !ok {
		return nil, domain.ErrCorrelationRuleNotFound
	}
	if sr.System {
		return nil, domain.ErrCorrelationRuleSystemOwner
	}

	abs := s.absPath(sr)
	if err := s.withStoreLock(func() error { return writeRuleFile(abs, rule) }); err != nil {
		return nil, err
	}

	sr.Rule = rule
	sr.Modified = fileModTime(abs)

	cp := *sr
	return &cp, nil
}

// Delete removes a user rule. System rules are read-only.
func (s *RuleStore) Delete(relPath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sr, ok := s.index[relPath]
	if !ok {
		return domain.ErrCorrelationRuleNotFound
	}
	if sr.System {
		return domain.ErrCorrelationRuleSystemOwner
	}

	if err := s.withStoreLock(func() error { return removeRuleFile(s.absPath(sr)) }); err != nil {
		return err
	}

	delete(s.index, relPath)
	for i, r := range s.rules {
		if r == sr {
			s.rules = append(s.rules[:i], s.rules[i+1:]...)
			break
		}
	}
	return nil
}

// SetEnabled records the answer for one tenant. The in-memory Enabled flag is
// deliberately left alone: it is shared by every tenant reading this store, and
// one customer switching a rule off must not switch it off for the others.
func (s *RuleStore) SetEnabled(tenantId, relPath string, enabled bool) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sr, ok := s.index[relPath]
	if !ok {
		return false, domain.ErrCorrelationRuleNotFound
	}
	if s.enabledFor(tenantId, relPath) == enabled {
		return false, nil
	}

	if err := s.withStoreLock(func() error {
		return s.writer.SetRuleDisabled(tenantId, ruleIdentity(relPath), !enabled)
	}); err != nil {
		return false, err
	}

	dir := s.userDir
	if sr.System {
		dir = s.systemDir
	}
	s.writer.PokeRuleReload(filepath.Join(dir, relPath))

	return true, nil
}

// AllRelPaths returns every known rule identity in load order (system first,
// then user).
func (s *RuleStore) AllRelPaths() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, 0, len(s.rules))
	for _, sr := range s.rules {
		out = append(out, sr.RelPath)
	}
	return out
}

func (s *RuleStore) ReadRuleBytes(relPath string) ([]byte, error) {
	s.mu.RLock()
	sr, ok := s.index[relPath]
	if !ok {
		s.mu.RUnlock()
		return nil, domain.ErrCorrelationRuleNotFound
	}
	abs := s.absPath(sr)
	s.mu.RUnlock()
	return os.ReadFile(abs)
}

func (s *RuleStore) DistinctValues(prop, value, tenantId string) []string {
	pick := propPicker(prop)
	if pick == nil {
		return []string{}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	seen := make(map[string]struct{})
	out := make([]string, 0)
	needle := strings.ToLower(value)
	for _, sr := range s.rules {
		if tenantId != "" && sr.TenantId != "" && sr.TenantId != tenantId {
			continue
		}
		v := pick(sr)
		if v == "" {
			continue
		}
		if needle != "" && !strings.Contains(strings.ToLower(v), needle) {
			continue
		}
		if _, dup := seen[v]; dup {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func propPicker(prop string) func(*domain.StoredRule) string {
	switch prop {
	case "rule_name":
		return func(sr *domain.StoredRule) string { return sr.Name }
	case "rule_category":
		return func(sr *domain.StoredRule) string { return sr.Category }
	case "rule_technique":
		return func(sr *domain.StoredRule) string { return sr.Technique }
	case "rule_adversary":
		return func(sr *domain.StoredRule) string { return string(sr.Adversary) }
	default:
		return nil
	}
}

func (s *RuleStore) absPath(sr *domain.StoredRule) string {
	dir := s.userDir
	if sr.System {
		dir = s.systemDir
	}
	return filepath.Join(dir, filepath.FromSlash(sr.RelPath))
}

// ── helpers ─────────────────────────────────────────────────────────────────

var slugNonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// slug turns a rule name into a filesystem-safe identifier.
func slug(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = slugNonAlnum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "rule"
	}
	return s
}

func fileModTime(path string) time.Time {
	if info, err := os.Stat(path); err == nil {
		return info.ModTime()
	}
	return time.Time{}
}

func containsStr(set []string, v string) bool {
	for _, x := range set {
		if x == v {
			return true
		}
	}
	return false
}

func containsInt(set []int, v int) bool {
	for _, x := range set {
		if x == v {
			return true
		}
	}
	return false
}

// anyStr reports whether any element of want is present in have.
func anyStr(want, have []string) bool {
	for _, w := range want {
		if containsStr(have, w) {
			return true
		}
	}
	return false
}

func containsAdversary(in []domain.AdversaryType, v domain.AdversaryType) bool {
	for _, a := range in {
		if a == v {
			return true
		}
	}
	return false
}

func writeRuleFile(path string, rule domain.Rule) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(rule)
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func readRuleFile(path string) (domain.Rule, error) {
	var rule domain.Rule
	data, err := os.ReadFile(path)
	if err != nil {
		return rule, err
	}
	if err := yaml.Unmarshal(data, &rule); err != nil {
		return rule, err
	}
	return rule, nil
}

// removeRuleFile deletes a rule file, tolerating a missing file.
func removeRuleFile(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func anyToRaw(v any) json.RawMessage {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

// enabledFor answers for one tenant, reading its own disabled list rather than
// the shared cache.
func (s *RuleStore) enabledFor(tenantId, relPath string) bool {
	return !s.writer.DisabledRuleSet(tenantId)[ruleIdentity(relPath)]
}
