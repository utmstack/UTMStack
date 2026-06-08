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

// ControlStore holds the control library across the system + user overlays,
// keyed by control ID (a user control overrides the system one — copy-on-write).
type ControlStore struct {
	systemDir string
	userDir   string
	mu        sync.RWMutex
	byID      map[string]*domain.Control
}

func NewControlStore(systemDir, userDir string) *ControlStore {
	return &ControlStore{systemDir: systemDir, userDir: userDir, byID: map[string]*domain.Control{}}
}

// Load rescans both overlays and rebuilds the in-memory index (user wins by ID).
func (s *ControlStore) Load() error {
	idx := map[string]*domain.Control{}
	for _, root := range []struct {
		dir    string
		system bool
	}{{s.systemDir, true}, {s.userDir, false}} {
		files, err := scanYAML(root.dir, root.system)
		if err != nil {
			return err
		}
		for _, f := range files {
			var c domain.Control
			if err := yaml.Unmarshal(f.data, &c); err != nil {
				continue
			}
			if strings.TrimSpace(c.ID) == "" {
				continue
			}
			c.RelPath = f.relPath
			c.System = f.system
			c.Enabled = f.enabled
			cp := c
			idx[c.ID] = &cp
		}
	}
	s.mu.Lock()
	s.byID = idx
	s.mu.Unlock()
	return nil
}

// All returns every control, sorted by ID.
func (s *ControlStore) All() []domain.Control {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Control, 0, len(s.byID))
	for _, c := range s.byID {
		out = append(out, *c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Get returns the control for the given ID (copy), or false.
func (s *ControlStore) Get(id string) (*domain.Control, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.byID[id]
	if !ok {
		return nil, false
	}
	cp := *c
	return &cp, true
}

// controlRelPath derives the user-overlay file path from family/id (the client's
// RelPath is ignored to prevent path traversal).
func controlRelPath(c *domain.Control) string {
	if c.Family != "" {
		return c.Family + "/" + c.ID + fileExt
	}
	return c.ID + fileExt
}

// Create writes a brand-new custom control to the user overlay. Fails if the id
// already exists (system or user).
func (s *ControlStore) Create(c domain.Control) (*domain.Control, error) {
	if !safeID(c.ID) {
		return nil, domain.ErrInvalidID
	}
	if _, ok := s.Get(c.ID); ok {
		return nil, domain.ErrControlExists
	}
	return s.writeUser(c)
}

// Update writes the control to the user overlay (copy-on-write): editing a system
// control creates a user override with the same id; the system file is untouched.
func (s *ControlStore) Update(c domain.Control) (*domain.Control, error) {
	if !safeID(c.ID) {
		return nil, domain.ErrInvalidID
	}
	if _, ok := s.Get(c.ID); !ok {
		return nil, domain.ErrControlNotFound
	}
	return s.writeUser(c)
}

func (s *ControlStore) writeUser(c domain.Control) (*domain.Control, error) {
	c.RelPath, c.System, c.Enabled = "", false, true // never persist store metadata
	data, err := yaml.Marshal(c)
	if err != nil {
		return nil, err
	}
	if err := atomicWrite(filepath.Join(s.userDir, controlRelPath(&c)), data); err != nil {
		return nil, err
	}
	if err := s.Load(); err != nil {
		return nil, err
	}
	out, _ := s.Get(c.ID)
	return out, nil
}

// Delete removes the user-overlay file for id. A vendor-only (system) control
// cannot be deleted — disable it instead. Deleting a user override of a system
// control reverts it to the vendor version.
func (s *ControlStore) Delete(id string) error {
	c, ok := s.Get(id)
	if !ok {
		return domain.ErrControlNotFound
	}
	rel := controlRelPath(c)
	userPath := filepath.Join(s.userDir, rel)
	if _, err := os.Stat(userPath); err != nil {
		if _, err2 := os.Stat(userPath + disabledSuffix); err2 != nil {
			return domain.ErrSystemOwner // no user file → vendor-only
		}
		userPath += disabledSuffix
	}
	if err := os.Remove(userPath); err != nil {
		return err
	}
	return s.Load()
}

// SetEnabled toggles the control's enabled state via the .disabled filename
// suffix, in its own overlay dir (system or user).
func (s *ControlStore) SetEnabled(id string, enabled bool) error {
	c, ok := s.Get(id)
	if !ok {
		return domain.ErrControlNotFound
	}
	dir := s.userDir
	if c.System {
		dir = s.systemDir
	}
	if err := setEnabledFile(dir, controlRelPath(c), enabled); err != nil {
		return err
	}
	return s.Load()
}
