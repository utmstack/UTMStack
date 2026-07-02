package usecase

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
)

type FlowListFilter struct {
	Page int // 0-based
	Size int

	Name   string // case-insensitive partial on flow name
	Search string // case-insensitive partial on flow name

	Active        *bool  // enabled state
	SystemOwner   *bool  // system overlay membership
	AgentPlatform string // exact match
}

type FlowStore struct {
	systemDir string
	userDir   string

	mu    sync.RWMutex
	flows []*StoredFlow
	index map[string]*StoredFlow // relPath -> flow
}

func NewFlowStore(systemDir, userDir string) *FlowStore {
	return &FlowStore{
		systemDir: systemDir,
		userDir:   userDir,
		index:     make(map[string]*StoredFlow),
	}
}

func (s *FlowStore) Load() error {
	flows := make([]*StoredFlow, 0, 64)
	index := make(map[string]*StoredFlow, 64)

	for _, ov := range []struct {
		dir    string
		system bool
	}{
		{s.systemDir, true},
		{s.userDir, false},
	} {
		loaded, err := loadFlowOverlay(ov.dir, ov.system)
		if err != nil {
			return err
		}
		for _, sf := range loaded {
			if prev, ok := index[sf.RelPath]; ok {
				*prev = *sf // user overlay overrides a system flow with the same relPath
				continue
			}
			index[sf.RelPath] = sf
			flows = append(flows, sf)
		}
	}

	s.mu.Lock()
	s.flows = flows
	s.index = index
	s.mu.Unlock()
	return nil
}

func loadFlowOverlay(dir string, system bool) ([]*StoredFlow, error) {
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

		out = append(out, &StoredFlow{
			Flow:     flow,
			RelPath:  filepath.ToSlash(rel),
			Modified: mod,
			system:   system,
			enabled:  enabled,
		})
		return nil
	})
	return out, err
}

func (s *FlowStore) Get(relPath string) *StoredFlow {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if sf, ok := s.index[relPath]; ok {
		cp := *sf
		return &cp
	}
	return nil
}

func (s *FlowStore) List(f FlowListFilter) ([]*StoredFlow, int) {
	s.mu.RLock()
	matched := make([]*StoredFlow, 0, len(s.flows))
	for _, sf := range s.flows {
		if flowMatches(sf, f) {
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

func (s *FlowStore) Create(flow Flow) (*StoredFlow, error) {
	relPath := flowSlug(flow.Name) + FlowFileExt

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.index[relPath]; exists {
		return nil, os.ErrExist
	}

	abs := filepath.Join(s.userDir, relPath)
	if err := writeFlowFile(abs, flow); err != nil {
		return nil, err
	}

	sf := &StoredFlow{
		Flow:     flow,
		RelPath:  relPath,
		Modified: flowModTime(abs),
		system:   false,
		enabled:  true,
	}
	s.index[relPath] = sf
	s.flows = append(s.flows, sf)

	cp := *sf
	return &cp, nil
}

func (s *FlowStore) Update(relPath string, flow Flow) (*StoredFlow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sf, ok := s.index[relPath]
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

func (s *FlowStore) Delete(relPath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sf, ok := s.index[relPath]
	if !ok {
		return ErrFlowNotFound
	}
	if sf.system {
		return ErrSystemFlowContent
	}

	if err := removeFlowFile(s.absPath(sf)); err != nil {
		return err
	}

	delete(s.index, relPath)
	for i, f := range s.flows {
		if f == sf {
			s.flows = append(s.flows[:i], s.flows[i+1:]...)
			break
		}
	}
	return nil
}

func (s *FlowStore) SetEnabled(relPath string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sf, ok := s.index[relPath]
	if !ok {
		return ErrFlowNotFound
	}
	if sf.enabled == enabled {
		return nil
	}

	from := s.absPath(sf)
	sf.enabled = enabled
	to := s.absPath(sf)
	if err := renameFlowFile(from, to); err != nil {
		sf.enabled = !enabled
		return err
	}
	return nil
}

func (s *FlowStore) absPath(sf *StoredFlow) string {
	dir := s.userDir
	if sf.system {
		dir = s.systemDir
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
