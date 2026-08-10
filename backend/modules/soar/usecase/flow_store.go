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

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
)

const flowReloadInterval = 2 * time.Second

type FlowListFilter struct {
	Page int // 0-based
	Size int

	Name   string // case-insensitive partial on flow name
	Search string // case-insensitive partial on flow name

	Active        *bool  // enabled state
	SystemOwner   *bool  // shipped with the product rather than written here
	AgentPlatform string // exact match
}

type flowKey struct {
	tenant  string
	relPath string
}

type FlowStore struct {
	systemDir string
	userDir   string

	mu      sync.RWMutex
	flows   []*domain.StoredFlow
	index   map[flowKey]*domain.StoredFlow
	enabled map[string]enabledSet // tenant → the flows they switched on
}

func NewFlowStore(systemDir, userDir string) *FlowStore {
	return &FlowStore{
		systemDir: systemDir,
		userDir:   userDir,
		index:     make(map[flowKey]*domain.StoredFlow),
		enabled:   make(map[string]enabledSet),
	}
}

func (s *FlowStore) Load() error {
	flows := make([]*domain.StoredFlow, 0, 128)
	index := make(map[flowKey]*domain.StoredFlow, 128)

	system, err := loadFlowDir(s.systemDir, "", true)
	if err != nil {
		return err
	}
	for _, sf := range system {
		k := flowKey{"", sf.RelPath}
		if _, ok := index[k]; ok {
			continue
		}
		index[k] = sf
		flows = append(flows, sf)
	}

	tenants, err := tenantFlowDirs(s.userDir)
	if err != nil {
		return err
	}
	enabled := make(map[string]enabledSet, len(tenants))
	for _, tenant := range tenants {
		dir := filepath.Join(s.userDir, tenant)
		owned, lerr := loadFlowDir(dir, tenant, false)
		if lerr != nil {
			return lerr
		}
		for _, sf := range owned {
			k := flowKey{tenant, sf.RelPath}
			if _, ok := index[k]; ok {
				continue
			}
			index[k] = sf
			flows = append(flows, sf)
		}
		on, derr := readEnabled(dir)
		if derr != nil {
			// Nothing on rather than everything on: an unreadable file must
			// never be the reason a command runs on somebody's machines.
			_ = catcher.Error("soar: unreadable enabled list, treating every flow as off", derr,
				map[string]any{"tenant": tenant})
			on = enabledSet{}
		}
		enabled[tenant] = on
	}

	s.mu.Lock()
	s.flows, s.index, s.enabled = flows, index, enabled
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

func tenantFlowDirs(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

func loadFlowDir(dir, tenant string, system bool) ([]*domain.StoredFlow, error) {
	var out []*domain.StoredFlow

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return out, nil
	}

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, FlowFileExt) {
			return nil
		}
		if filepath.Base(path) == EnabledFileName {
			return nil
		}

		flow, ferr := readFlowFile(path)
		if ferr != nil {
			_ = catcher.Error("soar: skipping unparsable flow file", ferr, map[string]any{"path": path})
			return nil // skip the file rather than fail the whole load
		}

		rel, ferr := filepath.Rel(dir, path)
		if ferr != nil {
			return nil
		}

		var mod time.Time
		if info, ierr := d.Info(); ierr == nil {
			mod = info.ModTime()
		}

		out = append(out, &domain.StoredFlow{
			Flow:     flow,
			RelPath:  filepath.ToSlash(rel),
			Modified: mod,
			System:   system,
			Tenant:   tenant,
		})
		return nil
	})
	return out, err
}

func (s *FlowStore) Get(tenant, relPath string) *domain.StoredFlow {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lookup(tenant, relPath)
}

func (s *FlowStore) lookup(tenant, relPath string) *domain.StoredFlow {
	sf, ok := s.index[flowKey{tenant, relPath}]
	if !ok {
		sf, ok = s.index[flowKey{"", relPath}]
	}
	if !ok {
		return nil
	}
	return s.viewFor(tenant, sf)
}

func (s *FlowStore) viewFor(tenant string, sf *domain.StoredFlow) *domain.StoredFlow {
	cp := *sf
	cp.Enabled = s.enabled[tenant][sf.RelPath]
	return &cp
}

func (s *FlowStore) List(tenant string, f FlowListFilter) ([]*domain.StoredFlow, int) {
	s.mu.RLock()
	seen := make(map[string]bool, len(s.flows))
	matched := make([]*domain.StoredFlow, 0, len(s.flows))
	for _, sf := range s.flows {
		if sf.Tenant != "" && sf.Tenant != tenant {
			continue
		}
		if sf.Tenant == "" && s.index[flowKey{tenant, sf.RelPath}] != nil {
			continue // the tenant wrote their own version of this one
		}
		if seen[sf.RelPath] {
			continue
		}
		// The view has to be built before matching: Active filters on a value
		// that only exists once the tenant's list is applied.
		cp := s.viewFor(tenant, sf)
		if flowMatches(cp, f) {
			seen[sf.RelPath] = true
			matched = append(matched, cp)
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
		return []*domain.StoredFlow{}, total
	}
	end := start + f.Size
	if end > total {
		end = total
	}
	return matched[start:end], total
}

func flowMatches(sf *domain.StoredFlow, f FlowListFilter) bool {
	if f.Name != "" && !strings.Contains(strings.ToLower(sf.Name), strings.ToLower(f.Name)) {
		return false
	}
	if f.Search != "" && !strings.Contains(strings.ToLower(sf.Name), strings.ToLower(f.Search)) {
		return false
	}
	if f.Active != nil && sf.Enabled != *f.Active {
		return false
	}
	if f.SystemOwner != nil && sf.System != *f.SystemOwner {
		return false
	}
	if f.AgentPlatform != "" && sf.AgentPlatform != f.AgentPlatform {
		return false
	}
	return true
}

func (s *FlowStore) Create(tenant string, flow domain.Flow) (*domain.StoredFlow, error) {
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

	sf := &domain.StoredFlow{
		Flow:     flow,
		RelPath:  relPath,
		Modified: flowModTime(abs),
		System:   false,
		Tenant:   tenant,
	}
	s.index[flowKey{tenant, relPath}] = sf
	s.flows = append(s.flows, sf)

	return s.viewFor(tenant, sf), nil
}

func (s *FlowStore) Update(tenant, relPath string, flow domain.Flow) (*domain.StoredFlow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sf, ok := s.index[flowKey{tenant, relPath}]
	if !ok {
		if _, shipped := s.index[flowKey{"", relPath}]; shipped {
			return nil, domain.ErrSystemFlowContent
		}
		return nil, domain.ErrFlowNotFound
	}

	abs := s.absPath(sf)
	if err := writeFlowFile(abs, flow); err != nil {
		return nil, err
	}

	sf.Flow = flow
	sf.Modified = flowModTime(abs)

	return s.viewFor(tenant, sf), nil
}

func (s *FlowStore) Delete(tenant, relPath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sf, ok := s.index[flowKey{tenant, relPath}]
	if !ok {
		if _, shipped := s.index[flowKey{"", relPath}]; shipped {
			return domain.ErrSystemFlowContent
		}
		return domain.ErrFlowNotFound
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

	if s.lookup(tenant, relPath) == nil {
		return domain.ErrFlowNotFound
	}

	next := make(enabledSet, len(s.enabled[tenant])+1)
	for k, v := range s.enabled[tenant] {
		next[k] = v
	}
	if enabled {
		next[relPath] = true
	} else {
		delete(next, relPath)
	}

	dir := filepath.Join(s.userDir, tenant)
	if err := s.withFlowLock(func() error { return writeEnabled(dir, next) }); err != nil {
		return err
	}
	s.enabled[tenant] = next
	return nil
}

func (s *FlowStore) absPath(sf *domain.StoredFlow) string {
	return filepath.Join(s.userDir, sf.Tenant, filepath.FromSlash(sf.RelPath))
}

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
