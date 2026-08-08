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

type FrameworkStore struct {
	systemDir string
	userRoot  string

	mu     sync.RWMutex
	system map[string]*domain.Framework
	user   map[string]map[string]*domain.Framework // tenant → key → framework
}

func NewFrameworkStore(systemDir, userRoot string) *FrameworkStore {
	return &FrameworkStore{
		systemDir: systemDir,
		userRoot:  userRoot,
		system:    map[string]*domain.Framework{},
		user:      map[string]map[string]*domain.Framework{},
	}
}

func (s *FrameworkStore) Load() error {
	sys, err := loadFrameworks(s.systemDir)
	if err != nil {
		return err
	}
	users := map[string]map[string]*domain.Framework{}
	for _, tid := range tenantDirs(s.userRoot) {
		m, err := loadFrameworks(filepath.Join(s.userRoot, tid))
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

func loadFrameworks(dir string) (map[string]*domain.Framework, error) {
	out := map[string]*domain.Framework{}
	files, err := scanYAML(dir, false)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		var fw domain.Framework
		if err := yaml.Unmarshal(f.data, &fw); err != nil || fw.Key == "" {
			continue
		}
		ff := fw
		out[fw.Key] = &ff
	}
	return out, nil
}

func (s *FrameworkStore) Get(ctx context.Context, key string) (*domain.Framework, bool) {
	tid, err := tenantDir(ctx)
	if err != nil {
		return nil, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if f, ok := s.user[tid][key]; ok {
		ff := *f
		return &ff, true
	}
	if f, ok := s.system[key]; ok {
		ff := *f
		return &ff, true
	}
	return nil, false
}

func (s *FrameworkStore) All(ctx context.Context) []domain.Framework {
	tid, err := tenantDir(ctx)
	if err != nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	merged := make(map[string]*domain.Framework, len(s.system)+len(s.user[tid]))
	for k, f := range s.system {
		merged[k] = f
	}
	for k, f := range s.user[tid] {
		merged[k] = f
	}
	out := make([]domain.Framework, 0, len(merged))
	for _, f := range merged {
		out = append(out, *f)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (s *FrameworkStore) IsSystem(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.system[key]
	return ok
}

func (s *FrameworkStore) Create(ctx context.Context, f domain.Framework) (*domain.Framework, error) {
	if !safeID(f.Key) {
		return nil, domain.ErrInvalidID
	}
	if _, ok := s.Get(ctx, f.Key); ok {
		return nil, domain.ErrFrameworkExists
	}
	return s.writeUser(ctx, f)
}

// Update refuses a vendor framework for the same reason controls do: the tenant
// layer sits beside the shipped catalogue, it does not fork it.
func (s *FrameworkStore) Update(ctx context.Context, f domain.Framework) (*domain.Framework, error) {
	if !safeID(f.Key) {
		return nil, domain.ErrInvalidID
	}
	if _, ok := s.Get(ctx, f.Key); !ok {
		return nil, domain.ErrFrameworkNotFound
	}
	if s.IsSystem(f.Key) {
		return nil, domain.ErrSystemOwner
	}
	return s.writeUser(ctx, f)
}

func (s *FrameworkStore) writeUser(ctx context.Context, f domain.Framework) (*domain.Framework, error) {
	tid, err := tenantDir(ctx)
	if err != nil {
		return nil, err
	}
	data, err := yaml.Marshal(f)
	if err != nil {
		return nil, err
	}
	if err := atomicWrite(filepath.Join(s.userRoot, tid, f.Key+fileExt), data); err != nil {
		return nil, err
	}
	if err := s.reloadTenant(tid); err != nil {
		return nil, err
	}
	out, _ := s.Get(ctx, f.Key)
	return out, nil
}

func (s *FrameworkStore) Delete(ctx context.Context, key string) error {
	if !safeID(key) {
		return domain.ErrInvalidID
	}
	tid, err := tenantDir(ctx)
	if err != nil {
		return err
	}
	if _, ok := s.Get(ctx, key); !ok {
		return domain.ErrFrameworkNotFound
	}
	path := filepath.Join(s.userRoot, tid, key+fileExt)
	if _, err := os.Stat(path); err != nil {
		return domain.ErrSystemOwner
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	return s.reloadTenant(tid)
}

func (s *FrameworkStore) reloadTenant(tid string) error {
	m, err := loadFrameworks(filepath.Join(s.userRoot, tid))
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.user[tid] = m
	s.mu.Unlock()
	return nil
}
