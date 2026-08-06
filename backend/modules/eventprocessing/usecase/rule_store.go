package usecase

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
)

// reloadInterval bounds how stale one backend replica's in-memory rule/filter
// cache can get relative to another replica's writes to the same shared
// directories. A short poll is simpler and more robust here than a recursive
// filesystem watch, since rules/filters are organized in nested subfolders.
const reloadInterval = 10 * time.Second

type RuleListFilter struct {
	Page int // 0-based
	Size int

	Name   string // case-insensitive partial on rule name
	Search string // case-insensitive partial on rule name (general text)

	Active      *bool // enabled state
	SystemOwner *bool // system overlay membership

	// TenantId, when set, restricts results to that tenant's own rules —
	// empty string means no filter (system + global + every tenant's rules).
	TenantId string

	Categories  []string // any-of
	Adversaries []string // any-of
	Techniques  []string // any-of
	DataTypes   []string // any-of (rule must reference at least one)

	Confidentiality []int // any-of
	Integrity       []int // any-of
	Availability    []int // any-of

	InitDate time.Time // Modified >= InitDate (when non-zero)
	EndDate  time.Time // Modified <= EndDate (when non-zero)
}

type RuleStore struct {
	systemDir string
	userDir   string
	writer    *pipelineWriter

	mu    sync.RWMutex
	rules []*StoredRule          // load order (system first, then user)
	index map[string]*StoredRule // relPath -> rule
}

func NewRuleStore(systemDir, userDir string, writer *pipelineWriter) *RuleStore {
	return &RuleStore{
		systemDir: systemDir,
		userDir:   userDir,
		writer:    writer,
		index:     make(map[string]*StoredRule),
	}
}

// ruleIdentity derives the same identity the event processor computes for a
// rule file: the filename without directory or extension (see
// plugins/cel/main.go's ruleIdentity). This is what tenants.yaml's
// disabledRules entries must match, so a rule's identity here is intentionally
// NOT its relPath — two vendors are free to share a relPath's basename only if
// their filenames actually differ (enforced by convention in definitions/rules).
func ruleIdentity(relPath string) string {
	base := filepath.Base(relPath)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// Load (re)reads both overlays into memory, replacing the current contents.
// Enabled state comes from tenants.yaml's disabledRules (via the shared
// pipelineWriter), not from any per-file marker.
func (s *RuleStore) Load() error {
	disabled := s.writer.DisabledRuleSet()

	rules := make([]*StoredRule, 0, 256)
	index := make(map[string]*StoredRule, 256)

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

// Watch periodically reloads the rule overlays so this replica picks up
// rules/enable-state written by another backend replica sharing the same
// directories. Intended to run as its own goroutine for the lifetime of ctx.
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

// loadOverlay walks one overlay directory and parses every rule file in it.
// disabled is keyed by ruleIdentity (see Load), not by relPath.
func loadOverlay(dir string, system bool, disabled map[string]bool) ([]*StoredRule, error) {
	var out []*StoredRule

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

		out = append(out, &StoredRule{
			Rule:     rule,
			RelPath:  rel,
			Modified: mod,
			system:   system,
			enabled:  !disabled[ruleIdentity(rel)],
		})
		return nil
	})
	return out, err
}

// findByRelPath returns the rule for relPath or nil. Caller-facing reads use
// the read lock.
func (s *RuleStore) findByRelPath(relPath string) *StoredRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if sr, ok := s.index[relPath]; ok {
		cp := *sr
		return &cp
	}
	return nil
}

func (s *RuleStore) FindByName(name string) *StoredRule {
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

// List applies the filter, sorts by name and paginates. It returns the page
// plus the total match count (pre-pagination).
func (s *RuleStore) List(f RuleListFilter) ([]*StoredRule, int) {
	s.mu.RLock()
	matched := make([]*StoredRule, 0, len(s.rules))
	for _, sr := range s.rules {
		if ruleMatches(sr, f) {
			cp := *sr
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
		return []*StoredRule{}, total
	}
	end := start + f.Size
	if end > total {
		end = total
	}
	return matched[start:end], total
}

func ruleMatches(sr *StoredRule, f RuleListFilter) bool {
	if f.TenantId != "" && sr.TenantId != f.TenantId {
		return false
	}
	if f.Name != "" && !strings.Contains(strings.ToLower(sr.Name), strings.ToLower(f.Name)) {
		return false
	}
	if f.Search != "" && !strings.Contains(strings.ToLower(sr.Name), strings.ToLower(f.Search)) {
		return false
	}
	if f.Active != nil && sr.enabled != *f.Active {
		return false
	}
	if f.SystemOwner != nil && sr.system != *f.SystemOwner {
		return false
	}
	if len(f.Categories) > 0 && !containsStr(f.Categories, sr.Category) {
		return false
	}
	if len(f.Adversaries) > 0 && !containsStr(f.Adversaries, sr.Adversary) {
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

// Create writes a new rule into the user overlay. The relPath is derived from
// the rule name; a collision is reported as an error.
// Create writes a new rule into the user overlay. tenantId scopes the rule to
// one tenant — empty means a global user rule, placed exactly as before this
// field existed. A tenant-owned rule lives in its own subfolder (userDir/
// <tenantId>/), and its filename is suffixed with the tenant id too: identity
// (see ruleIdentity) is basename-only, so without the suffix two tenants
// picking the same rule name would collide on the exact same identity the
// engine uses for enable/disable and error-tracking.
func (s *RuleStore) Create(rule Rule, tenantId string) (*StoredRule, error) {
	rule.TenantId = tenantId

	var relPath string
	if tenantId == "" {
		relPath = slug(rule.Name) + RuleFileExt
	} else {
		relPath = filepath.ToSlash(filepath.Join(tenantId, slug(rule.Name)+"-"+tenantId+RuleFileExt))
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.index[relPath]; exists {
		return nil, os.ErrExist
	}

	abs := filepath.Join(s.userDir, filepath.FromSlash(relPath))
	if err := writeRuleFile(abs, rule); err != nil {
		return nil, err
	}

	sr := &StoredRule{
		Rule:     rule,
		RelPath:  relPath,
		Modified: fileModTime(abs),
		system:   false,
		enabled:  true,
	}
	s.index[relPath] = sr
	s.rules = append(s.rules, sr)

	cp := *sr
	return &cp, nil
}

// Update overwrites an existing user rule's content. System rules are
// read-only.
func (s *RuleStore) Update(relPath string, rule Rule) (*StoredRule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sr, ok := s.index[relPath]
	if !ok {
		return nil, ErrRuleNotFound
	}
	if sr.system {
		return nil, ErrSystemRuleContent
	}

	abs := s.absPath(sr)
	if err := writeRuleFile(abs, rule); err != nil {
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
		return ErrRuleNotFound
	}
	if sr.system {
		return ErrSystemRuleContent
	}

	if err := removeRuleFile(s.absPath(sr)); err != nil {
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

// SetEnabled toggles a rule's enabled state. This is recorded in tenants.yaml
// (disabledRules, keyed by ruleIdentity) rather than on the file itself, so it
// works identically for system and user rules (disabling is the one mutation
// allowed on a system rule) and never touches file content.
//
// The returned bool is the authoritative "did this call actually flip the
// state" signal: false when the rule was already in the requested state (a
// no-op) or when the call failed, true only when this call performed the
// transition. Callers that need to know whether THEY caused a state change
// (e.g. to decide whether to notify) must use this return value rather than
// re-deriving it from a value read before the call.
func (s *RuleStore) SetEnabled(relPath string, enabled bool) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sr, ok := s.index[relPath]
	if !ok {
		return false, ErrRuleNotFound
	}
	if sr.enabled == enabled {
		return false, nil
	}

	if err := s.writer.SetRuleDisabled(ruleIdentity(relPath), !enabled); err != nil {
		return false, err
	}
	sr.enabled = enabled

	dir := s.userDir
	if sr.system {
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

// ReadRuleBytes returns the raw on-disk YAML for a rule, preserving comments
// and formatting. Only rules present in the index are readable, so relPath is
// safe against traversal (it must match a known entry).
func (s *RuleStore) ReadRuleBytes(relPath string) ([]byte, error) {
	s.mu.RLock()
	sr, ok := s.index[relPath]
	if !ok {
		s.mu.RUnlock()
		return nil, ErrRuleNotFound
	}
	abs := s.absPath(sr)
	s.mu.RUnlock()
	return os.ReadFile(abs)
}

// DistinctValues returns the distinct values of a rule property, optionally
// filtered to those containing `value` (case-insensitive). prop is a legacy
// column name (rule_name, rule_category, rule_technique, rule_adversary).
func (s *RuleStore) DistinctValues(prop, value string) []string {
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

func propPicker(prop string) func(*StoredRule) string {
	switch prop {
	case "rule_name":
		return func(sr *StoredRule) string { return sr.Name }
	case "rule_category":
		return func(sr *StoredRule) string { return sr.Category }
	case "rule_technique":
		return func(sr *StoredRule) string { return sr.Technique }
	case "rule_adversary":
		return func(sr *StoredRule) string { return sr.Adversary }
	default:
		return nil
	}
}

// absPath resolves a rule's on-disk path. Enabled state lives in tenants.yaml,
// not in the filename, so this is just the overlay dir joined with relPath.
func (s *RuleStore) absPath(sr *StoredRule) string {
	dir := s.userDir
	if sr.system {
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
