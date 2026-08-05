package usecase

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
)

const flowReloadInterval = 2 * time.Second

type FlowListFilter struct {
	Page int // 0-based
	Size int

	Name   string // case-insensitive partial on flow name
	Search string // case-insensitive partial on flow name

	Active        *bool  // enabled state
	SystemOwner   *bool  // system overlay membership
	AgentPlatform string // exact match
}

type flowKey struct {
	tenant  string
	relPath string
}

type FlowStore struct {
	systemDir string
	userDir   string

	mu    sync.RWMutex
	flows []*StoredFlow
	index map[flowKey]*StoredFlow
}

func NewFlowStore(systemDir, userDir string) *FlowStore {
	return &FlowStore{
		systemDir: systemDir,
		userDir:   userDir,
		index:     make(map[flowKey]*StoredFlow),
	}
}

func (s *FlowStore) Load() error {
	flows := make([]*StoredFlow, 0, 64)
	index := make(map[flowKey]*StoredFlow, 64)

	loaded, err := loadFlowOverlay(s.userDir)
	if err != nil {
		return err
	}
	for _, sf := range loaded {
		k := flowKey{sf.tenant, sf.RelPath}
		if prev, ok := index[k]; ok {
			*prev = *sf
			continue
		}
		index[k] = sf
		flows = append(flows, sf)
	}

	s.mu.Lock()
	s.flows = flows
	s.index = index
	s.mu.Unlock()
	return nil
}

func (s *FlowStore) Watch(ctx context.Context) {
	ticker := time.NewTicker(flowReloadInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.Load(); err != nil {
				_ = catcher.Error("soar: periodic flow reload failed", err, nil)
			}
		}
	}
}

func loadFlowOverlay(dir string) ([]*StoredFlow, error) {
	var out []*StoredFlow

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

		enabled := true
		logical := path
		if strings.HasSuffix(path, DisabledSuffix) {
			enabled = false
			logical = strings.TrimSuffix(path, DisabledSuffix)
		}
		if !strings.HasSuffix(logical, FlowFileExt) {
			return nil
		}

		flow, ferr := readFlowFile(path)
		if ferr != nil {
			_ = catcher.Error("soar: skipping unparsable flow file", ferr, map[string]any{"path": path})
			return nil // skip unparsable files rather than failing the whole load
		}

		rel, ferr := filepath.Rel(dir, logical)
		if ferr != nil {
			return nil
		}

		var mod time.Time
		if info, ierr := d.Info(); ierr == nil {
			mod = info.ModTime()
		}

		relSlash := filepath.ToSlash(rel)
		i := strings.Index(relSlash, "/")
		if i <= 0 {
			return nil // a file naming no tenant belongs to nobody
		}
		tenant := relSlash[:i]
		relSlash = relSlash[i+1:]

		system := false
		if after, ok := strings.CutPrefix(relSlash, SystemSubdir+"/"); ok {
			system, relSlash = true, after
		}

		out = append(out, &StoredFlow{
			Flow:     flow,
			RelPath:  relSlash,
			Modified: mod,
			system:   system,
			enabled:  enabled,
			tenant:   tenant,
		})
		return nil
	})
	return out, err
}

func (s *FlowStore) Get(tenant, relPath string) *StoredFlow {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lookup(tenant, relPath)
}

func (s *FlowStore) lookup(tenant, relPath string) *StoredFlow {
	if sf, ok := s.index[flowKey{tenant, relPath}]; ok {
		cp := *sf
		return &cp
	}
	if sf, ok := s.index[flowKey{"", relPath}]; ok {
		cp := *sf
		return &cp
	}
	return nil
}

func (s *FlowStore) List(tenant string, f FlowListFilter) ([]*StoredFlow, int) {
	s.mu.RLock()
	seen := make(map[string]bool, len(s.flows))
	matched := make([]*StoredFlow, 0, len(s.flows))
	for _, sf := range s.flows {
		if sf.tenant != "" && sf.tenant != tenant {
			continue
		}
		if sf.tenant == "" && s.index[flowKey{tenant, sf.RelPath}] != nil {
			continue // the tenant overrode this system flow
		}
		if seen[sf.RelPath] {
			continue
		}
		if flowMatches(sf, f) {
			seen[sf.RelPath] = true
			cp := *sf
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
		return []*StoredFlow{}, total
	}
	end := start + f.Size
	if end > total {
		end = total
	}
	return matched[start:end], total
}

func flowMatches(sf *StoredFlow, f FlowListFilter) bool {
	if f.Name != "" && !strings.Contains(strings.ToLower(sf.Name), strings.ToLower(f.Name)) {
		return false
	}
	if f.Search != "" && !strings.Contains(strings.ToLower(sf.Name), strings.ToLower(f.Search)) {
		return false
	}
	if f.Active != nil && sf.enabled != *f.Active {
		return false
	}
	if f.SystemOwner != nil && sf.system != *f.SystemOwner {
		return false
	}
	if f.AgentPlatform != "" && sf.AgentPlatform != f.AgentPlatform {
		return false
	}
	return true
}

func (s *FlowStore) Create(tenant string, flow Flow) (*StoredFlow, error) {
	relPath := flowSlug(flow.Name) + FlowFileExt

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.index[flowKey{tenant, relPath}]; exists {
		return nil, os.ErrExist
	}

	abs := filepath.Join(s.userDir, tenant, relPath)
	if err := s.withFlowLock(func() error {
		if s.flowExistsOnDisk(tenant, relPath) {
			return os.ErrExist
		}
		return writeFlowFile(abs, flow)
	}); err != nil {
		return nil, err
	}

	sf := &StoredFlow{
		Flow:     flow,
		RelPath:  relPath,
		Modified: flowModTime(abs),
		system:   false,
		enabled:  true,
		tenant:   tenant,
	}
	s.index[flowKey{tenant, relPath}] = sf
	s.flows = append(s.flows, sf)

	cp := *sf
	return &cp, nil
}

func (s *FlowStore) Update(tenant, relPath string, flow Flow) (*StoredFlow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sf, ok := s.index[flowKey{tenant, relPath}]
	if !ok {
		return nil, ErrFlowNotFound
	}
	if sf.system {
		return nil, ErrSystemFlowContent
	}

	abs := s.absPath(sf)
	if err := writeFlowFile(abs, flow); err != nil {
		return nil, err
	}

	sf.Flow = flow
	sf.Modified = flowModTime(abs)

	cp := *sf
	return &cp, nil
}

func (s *FlowStore) Delete(tenant, relPath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sf, ok := s.index[flowKey{tenant, relPath}]
	if !ok {
		return ErrFlowNotFound
	}
	if sf.system {
		return ErrSystemFlowContent
	}

	if err := removeFlowFile(s.absPath(sf)); err != nil {
		return err
	}

	delete(s.index, flowKey{tenant, relPath})
	for i, f := range s.flows {
		if f == sf {
			s.flows = append(s.flows[:i], s.flows[i+1:]...)
			break
		}
	}
	return nil
}

func (s *FlowStore) SetEnabled(tenant, relPath string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sf, ok := s.index[flowKey{tenant, relPath}]
	if !ok {
		return ErrFlowNotFound
	}
	if sf.enabled == enabled {
		return nil
	}

	from := s.absPath(sf)
	sf.enabled = enabled
	to := s.absPath(sf)
	if err := s.withFlowLock(func() error { return renameFlowFile(from, to) }); err != nil {
		sf.enabled = !enabled
		return err
	}
	return nil
}

func (s *FlowStore) absPath(sf *StoredFlow) string {
	dir := filepath.Join(s.userDir, sf.tenant)
	if sf.system {
		dir = filepath.Join(dir, SystemSubdir)
	}
	p := filepath.Join(dir, filepath.FromSlash(sf.RelPath))
	if !sf.enabled {
		p += DisabledSuffix
	}
	return p
}

// ── helpers ─────────────────────────────────────────────────────────────────

var flowSlugNonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

func flowSlug(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = flowSlugNonAlnum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "flow"
	}
	return s
}

func flowModTime(path string) time.Time {
	if info, err := os.Stat(path); err == nil {
		return info.ModTime()
	}
	return time.Time{}
}
