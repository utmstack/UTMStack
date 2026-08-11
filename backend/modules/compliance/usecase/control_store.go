package usecase

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
)

type ControlStore struct {
	systemDir string
	userRoot  string

	mu     sync.RWMutex
	system map[string]*domain.Control
	user   map[string]map[string]*domain.Control // tenant → id → control
}

func NewControlStore(systemDir, userRoot string) *ControlStore {
	return &ControlStore{
		systemDir: systemDir,
		userRoot:  userRoot,
		system:    map[string]*domain.Control{},
		user:      map[string]map[string]*domain.Control{},
	}
}

func (s *ControlStore) Load() error {
	sys, err := loadControls(s.systemDir, true)
	if err != nil {
		return err
	}
	users := map[string]map[string]*domain.Control{}
	for _, tid := range tenantDirs(s.userRoot) {
		m, err := loadControls(filepath.Join(s.userRoot, tid), false)
		if err != nil {
			return err
		}
		users[tid] = m
	}
	s.mu.Lock()
	s.system, s.user = sys, users
	s.mu.Unlock()
	return nil
}

func loadControls(dir string, system bool) (map[string]*domain.Control, error) {
	out := map[string]*domain.Control{}
	files, err := scanYAML(dir, system)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		var c domain.Control
		if err := yaml.Unmarshal(f.data, &c); err != nil || c.ID == "" {
			continue
		}
		cc := c
		out[c.ID] = &cc
	}
	return out, nil
}

func (s *ControlStore) Get(ctx context.Context, id string) (*domain.Control, bool) {
	tid, err := tenantDir(ctx)
	if err != nil {
		return nil, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if c, ok := s.user[tid][id]; ok {
		cc := *c
		return &cc, true
	}
	if c, ok := s.system[id]; ok {
		cc := *c
		return &cc, true
	}
	return nil, false
}

func (s *ControlStore) All(ctx context.Context) []domain.Control {
	tid, err := tenantDir(ctx)
	if err != nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	merged := make(map[string]*domain.Control, len(s.system)+len(s.user[tid]))
	for id, c := range s.system {
		merged[id] = c
	}
	for id, c := range s.user[tid] {
		merged[id] = c
	}
	out := make([]domain.Control, 0, len(merged))
	for _, c := range merged {
		out = append(out, *c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (s *ControlStore) Create(ctx context.Context, c domain.Control) (*domain.Control, error) {
	if !safeID(c.ID) {
		return nil, domain.ErrInvalidID
	}
	if _, ok := s.Get(ctx, c.ID); ok {
		return nil, domain.ErrControlExists
	}
	return s.writeUser(ctx, c)
}

// Update writes to the tenant's overlay, and refuses to touch a vendor control.
//
// The tenant layer is additive: a tenant builds its own controls beside the
// shipped library rather than forking it. An override would fragment the
// crosswalk — the same id meaning different things per tenant — and would put
// the copy beyond the reach of the next release's corrections.
func (s *ControlStore) Update(ctx context.Context, c domain.Control) (*domain.Control, error) {
	if !safeID(c.ID) {
		return nil, domain.ErrInvalidID
	}
	if _, ok := s.Get(ctx, c.ID); !ok {
		return nil, domain.ErrControlNotFound
	}
	if s.IsSystem(c.ID) {
		return nil, domain.ErrSystemOwner
	}
	return s.writeUser(ctx, c)
}

func (s *ControlStore) writeUser(ctx context.Context, c domain.Control) (*domain.Control, error) {
	tid, err := tenantDir(ctx)
	if err != nil {
		return nil, err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return nil, err
	}
	if err := withTenantLock(filepath.Join(s.userRoot, tid), func() error {
		return atomicWrite(filepath.Join(s.userRoot, tid, c.ID+fileExt), data)
	}); err != nil {
		return nil, err
	}
	if err := s.reloadTenant(tid); err != nil {
		return nil, err
	}
	out, _ := s.Get(ctx, c.ID)
	return out, nil
}

func (s *ControlStore) Delete(ctx context.Context, id string) error {
	if !safeID(id) {
		return domain.ErrInvalidID
	}
	tid, err := tenantDir(ctx)
	if err != nil {
		return err
	}
	if _, ok := s.Get(ctx, id); !ok {
		return domain.ErrControlNotFound
	}
	path := filepath.Join(s.userRoot, tid, id+fileExt)
	if err := withTenantLock(filepath.Join(s.userRoot, tid), func() error {
		if _, err := os.Stat(path); err != nil {
			return domain.ErrSystemOwner
		}
		return os.Remove(path)
	}); err != nil {
		return err
	}
	return s.reloadTenant(tid)
}

func (s *ControlStore) reloadTenant(tid string) error {
	m, err := loadControls(filepath.Join(s.userRoot, tid), false)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.user[tid] = m
	s.mu.Unlock()
	return nil
}

func (s *ControlStore) IsSystem(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.system[id]
	return ok
}
