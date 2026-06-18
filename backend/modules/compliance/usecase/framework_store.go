package usecase

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"gopkg.in/yaml.v3"
)

// FrameworkStore holds the framework checklists across system + user overlays,
// keyed by framework key (user wins).
type FrameworkStore struct {
	systemDir string
	userDir   string
	mu        sync.RWMutex
	byKey     map[string]*domain.Framework
}

func NewFrameworkStore(systemDir, userDir string) *FrameworkStore {
	return &FrameworkStore{systemDir: systemDir, userDir: userDir, byKey: map[string]*domain.Framework{}}
}

func (s *FrameworkStore) Load() error {
	idx := map[string]*domain.Framework{}
	for _, root := range []struct {
		dir    string
		system bool
	}{{s.systemDir, true}, {s.userDir, false}} {
		files, err := scanYAML(root.dir, root.system)
		if err != nil {
			return err
		}
		for _, f := range files {
			var fw domain.Framework
			if err := yaml.Unmarshal(f.data, &fw); err != nil {
				continue
			}
			if strings.TrimSpace(fw.Key) == "" {
				continue
			}
			fw.RelPath = f.relPath
			fw.System = f.system
			fw.Enabled = f.enabled
			cp := fw
			idx[fw.Key] = &cp
		}
	}
	s.mu.Lock()
	s.byKey = idx
	s.mu.Unlock()
	return nil
}

// All returns every framework, sorted by key.
func (s *FrameworkStore) All() []domain.Framework {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Framework, 0, len(s.byKey))
	for _, f := range s.byKey {
		out = append(out, *f)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

// Get returns the framework for the given key (copy), or false.
func (s *FrameworkStore) Get(key string) (*domain.Framework, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f, ok := s.byKey[key]
	if !ok {
		return nil, false
	}
	cp := *f
	return &cp, true
}

func frameworkRelPath(f *domain.Framework) string { return f.Key + fileExt }

// Create writes a brand-new custom framework to the user overlay. Fails if the
// key already exists (system or user).
func (s *FrameworkStore) Create(f domain.Framework) (*domain.Framework, error) {
	if !safeID(f.Key) {
		return nil, domain.ErrInvalidID
	}
	if _, ok := s.Get(f.Key); ok {
		return nil, domain.ErrFrameworkExists
	}
	return s.writeUser(f)
}

// Update writes the framework to the user overlay (copy-on-write).
func (s *FrameworkStore) Update(f domain.Framework) (*domain.Framework, error) {
	if !safeID(f.Key) {
		return nil, domain.ErrInvalidID
	}
	if _, ok := s.Get(f.Key); !ok {
		return nil, domain.ErrFrameworkNotFound
	}
	return s.writeUser(f)
}

func (s *FrameworkStore) writeUser(f domain.Framework) (*domain.Framework, error) {
	f.RelPath, f.System, f.Enabled = "", false, true
	data, err := yaml.Marshal(f)
	if err != nil {
		return nil, err
	}
	if err := atomicWrite(filepath.Join(s.userDir, frameworkRelPath(&f)), data); err != nil {
		return nil, err
	}
	if err := s.Load(); err != nil {
		return nil, err
	}
	out, _ := s.Get(f.Key)
	return out, nil
}

// Delete removes the user-overlay file for key. A vendor-only framework cannot be
// deleted — disable it instead.
func (s *FrameworkStore) Delete(key string) error {
	f, ok := s.Get(key)
	if !ok {
		return domain.ErrFrameworkNotFound
	}
	rel := frameworkRelPath(f)
	userPath := filepath.Join(s.userDir, rel)
	if _, err := os.Stat(userPath); err != nil {
		if _, err2 := os.Stat(userPath + disabledSuffix); err2 != nil {
			return domain.ErrSystemOwner
		}
		userPath += disabledSuffix
	}
	if err := os.Remove(userPath); err != nil {
		return err
	}
	return s.Load()
}

// SetEnabled toggles the framework's enabled state via the .disabled suffix.
func (s *FrameworkStore) SetEnabled(key string, enabled bool) error {
	f, ok := s.Get(key)
	if !ok {
		return domain.ErrFrameworkNotFound
	}
	dir := s.userDir
	if f.System {
		dir = s.systemDir
	}
	if err := setEnabledFile(dir, frameworkRelPath(f), enabled); err != nil {
		return err
	}
	return s.Load()
}
