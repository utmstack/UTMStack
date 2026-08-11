package repository

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"gopkg.in/yaml.v3"

	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/domain"
)

type genericPipelineFile struct {
	Pipeline []map[string]any `yaml:"pipeline"`
}

func filterTenantID(content []byte) string {
	var cfg genericPipelineFile
	if err := yaml.Unmarshal(content, &cfg); err != nil || len(cfg.Pipeline) == 0 {
		return ""
	}
	v, _ := cfg.Pipeline[0]["tenantId"].(string)
	return v
}

func withTenantID(content []byte, tenantId string) ([]byte, error) {
	var cfg genericPipelineFile
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, err
	}
	for i := range cfg.Pipeline {
		if tenantId == "" {
			delete(cfg.Pipeline[i], "tenantId")
		} else {
			cfg.Pipeline[i]["tenantId"] = tenantId
		}
	}
	return yaml.Marshal(cfg)
}

type PipelineStore struct {
	systemDir string
	userDir   string
	writer    connectors.EngineConfigRepository
	mu        sync.RWMutex
	filters   map[string]*domain.Pipeline // keyed by RelPath
}

func NewPipelineStore(systemDir, userDir string, writer connectors.EngineConfigRepository) *PipelineStore {
	return &PipelineStore{
		systemDir: systemDir,
		userDir:   userDir,
		writer:    writer,
		filters:   make(map[string]*domain.Pipeline),
	}
}

func (s *PipelineStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.filters = make(map[string]*domain.Pipeline)

	for _, root := range []struct {
		dir    string
		system bool
	}{{s.systemDir, true}, {s.userDir, false}} {
		if _, err := os.Stat(root.dir); os.IsNotExist(err) {
			continue
		}
		err := filepath.WalkDir(root.dir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			base := path
			active := true
			if strings.HasSuffix(base, DisabledSuffix) {
				base = strings.TrimSuffix(base, DisabledSuffix)
				active = false
			}
			if filepath.Ext(base) != PipelineFileExt {
				return nil
			}
			rel, err := filepath.Rel(root.dir, base)
			if err != nil {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			s.filters[rel] = &domain.Pipeline{
				RelPath:  rel,
				Content:  data,
				System:   root.system,
				Active:   active,
				TenantID: filterTenantID(data),
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Watch periodically reloads the filter overlays so this replica picks up
// filters/enable-state written by another backend replica sharing the same
// directories. Intended to run as its own goroutine for the lifetime of ctx.
func (s *PipelineStore) Watch(ctx context.Context) {
	ticker := time.NewTicker(reloadInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.Load(); err != nil {
				_ = catcher.Error("eventprocessing: periodic filter reload failed", err, nil)
			}
		}
	}
}

// List returns all filter entries (copy).
func (s *PipelineStore) List(tenantId string) []domain.Pipeline {
	disabled := s.writer.DisabledPipelineSet(tenantId)

	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Pipeline, 0, len(s.filters))
	for _, f := range s.filters {
		cp := *f
		// Active is per tenant, so it is stamped here rather than trusted from
		// the shared cache. A pipeline shipped switched off stays off for
		// everybody; a tenant can only switch off what it can see.
		cp.Active = cp.Active && !disabled[pipelineIdentity(cp.RelPath)]
		out = append(out, cp)
	}
	return out
}

// GetByRelPath returns the entry for the given relPath, or nil.
func (s *PipelineStore) GetByRelPath(relPath string) *domain.Pipeline {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f := s.filters[relPath]
	if f == nil {
		return nil
	}
	cp := *f
	return &cp
}

func (s *PipelineStore) withStoreLock(fn func() error) error {
	return withFileLock(filepath.Join(s.userDir, ".pipelines"), fn)
}

func (s *PipelineStore) Create(relPath string, content []byte, tenantId string) (*domain.Pipeline, error) {
	if tenantId != "" {
		injected, err := withTenantID(content, tenantId)
		if err != nil {
			return nil, fmt.Errorf("invalid filter content: %w", err)
		}
		content = injected

		base := filepath.Base(relPath)
		ext := filepath.Ext(base)
		name := strings.TrimSuffix(base, ext)
		relPath = filepath.ToSlash(filepath.Join(tenantId, name+"-"+tenantId+ext))
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	target := filepath.Join(s.userDir, filepath.FromSlash(relPath))
	if err := s.withStoreLock(func() error {
		if _, err := os.Stat(target); err == nil {
			return os.ErrExist
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return atomicWrite(target, content)
	}); err != nil {
		return nil, err
	}
	entry := &domain.Pipeline{RelPath: relPath, Content: content, System: false, Active: true, TenantID: tenantId}
	s.filters[relPath] = entry
	cp := *entry
	return &cp, nil
}

func (s *PipelineStore) Update(relPath string, content []byte) (*domain.Pipeline, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.filters[relPath]
	if !ok {
		return nil, domain.ErrPipelineNotFound
	}
	if existing.System {
		return nil, domain.ErrPipelineSystemOwner
	}
	target := filepath.Join(s.userDir, relPath)
	if err := s.withStoreLock(func() error { return atomicWrite(target, content) }); err != nil {
		return nil, err
	}
	existing.Content = content
	cp := *existing
	return &cp, nil
}

func (s *PipelineStore) Delete(relPath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.filters[relPath]
	if !ok {
		return domain.ErrPipelineNotFound
	}
	if existing.System {
		return domain.ErrPipelineSystemOwner
	}
	target := filepath.Join(s.userDir, relPath)
	_ = s.withStoreLock(func() error {
		// Try enabled path; fall back to disabled.
		if err := os.Remove(target); err != nil {
			_ = os.Remove(target + DisabledSuffix)
		}
		return nil
	})
	delete(s.filters, relPath)
	return nil
}

// SetEnabled records the answer for one tenant in tenants.yaml, the same place
// rules keep theirs.
//
// It used to rename the file to .disabled, which for a shipped pipeline is one
// copy shared by every tenant: one customer switching it off switched it off
// for all of them. The .disabled suffix still works on disk — it is how a
// pipeline ships switched off — but it is no longer how a tenant answers.
func (s *PipelineStore) SetEnabled(tenantId, relPath string, active bool) error {
	s.mu.RLock()
	_, ok := s.filters[relPath]
	s.mu.RUnlock()
	if !ok {
		return domain.ErrPipelineNotFound
	}
	return s.writer.SetPipelineDisabled(tenantId, pipelineIdentity(relPath), !active)
}

// pipelineIdentity is the name the engine matches on: the file's base name
// without its extension.
func pipelineIdentity(relPath string) string {
	base := filepath.Base(relPath)
	base = strings.TrimSuffix(base, DisabledSuffix)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// atomicWrite writes data to path via a temp file + rename (0600).
func atomicWrite(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
