package repository

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/utmstack/utmstack/backend/modules/integrations/connectors"
	"github.com/utmstack/utmstack/backend/modules/integrations/domain"
	"gopkg.in/yaml.v3"
)

// lockStaleness is how long a lock can sit untouched before we assume the
// writer crashed and steal it. Real writes are a few ms; anything past this is
// almost certainly a dead process.
const lockStaleness = 30 * time.Second

// withFileLock acquires an exclusive cross-process lock via a <path>.lock file,
// runs fn, then removes the lock. O_EXCL makes check-and-create atomic. If the
// existing lock is older than lockStaleness (writer crashed), it is stolen.
// ponytail: mtime-based staleness has a small race window if two processes
// steal simultaneously; upgrade to flock/fcntl if we ever see corruption.
func withFileLock(path string, fn func() error) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	lock := path + ".lock"
	for {
		f, err := os.OpenFile(lock, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			_ = f.Close()
			defer os.Remove(lock)
			return fn()
		}
		if !errors.Is(err, os.ErrExist) {
			return err
		}
		if info, statErr := os.Stat(lock); statErr == nil && time.Since(info.ModTime()) > lockStaleness {
			_ = os.Remove(lock)
			continue
		}
		time.Sleep(50 * time.Millisecond)
	}
}

var _ connectors.TenantRepository = (*TenantStore)(nil)

type TenantStore struct {
	dir string
	mu  sync.Mutex
}

func NewTenantStore(dir string) *TenantStore {
	return &TenantStore{dir: dir}
}

// pluginNames maps a configurable integration to the plugin that consumes its
// config. The runtime tenant file is system_plugins_<plugin>.yaml so the
// file-reading puller plugin can find it (note AWS_IAM_USER -> aws).
var pluginNames = map[string]string{
	"AWS_IAM_USER": "aws",
	"AZURE":        "azure",
	"GCP":          "gcp",
	"O365":         "o365",
	"BITDEFENDER":  "bitdefender",
	"CROWDSTRIKE":  "crowdstrike",
	"SOPHOS":       "sophos",
}

func pluginKey(module string) string {
	name, ok := pluginNames[strings.ToUpper(module)]
	if !ok {
		name = strings.ToLower(module)
	}
	return name
}

func tenantFileName(module string) string {
	return "system_plugins_" + pluginKey(module) + ".yaml"
}

func (s *TenantStore) path(module string) string {
	return filepath.Join(s.dir, tenantFileName(module))
}

// pluginsFile is the on-disk wrapper: plugins.<name>.tenants: [...].
type pluginsFile struct {
	Plugins map[string]pluginEntry `yaml:"plugins"`
}

type pluginEntry struct {
	Tenants []domain.Tenant `yaml:"tenants"`
}

func (s *TenantStore) SetActiveByModule(_ context.Context, module string, active bool) error {
	if active {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	return withFileLock(s.path(module), func() error {
		if err := os.Remove(s.path(module)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	})
}

func (s *TenantStore) Load(module string) ([]domain.Tenant, error) {
	data, err := os.ReadFile(s.path(module))
	if errors.Is(err, os.ErrNotExist) {
		return []domain.Tenant{}, nil
	}
	if err != nil {
		return nil, err
	}
	var pf pluginsFile
	if err := yaml.Unmarshal(data, &pf); err != nil {
		return nil, err
	}
	entry, ok := pf.Plugins[pluginKey(module)]
	if !ok {
		return []domain.Tenant{}, nil
	}
	return entry.Tenants, nil
}

func (s *TenantStore) Save(module string, tenants []domain.Tenant) error {
	return withFileLock(s.path(module), func() error {
		return s.saveLocked(module, tenants)
	})
}

// saveLocked writes the file assuming the caller already holds the file lock.
func (s *TenantStore) saveLocked(module string, tenants []domain.Tenant) error {
	key := pluginKey(module)
	pf := pluginsFile{Plugins: map[string]pluginEntry{key: {Tenants: tenants}}}
	data, err := yaml.Marshal(pf)
	if err != nil {
		return err
	}
	dst := s.path(module)
	tmp := dst + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}

func (s *TenantStore) Upsert(module string, t domain.Tenant) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return withFileLock(s.path(module), func() error {
		tenants, err := s.Load(module)
		if err != nil {
			return err
		}
		for i := range tenants {
			if tenants[i].Name == t.Name {
				tenants[i] = t
				return s.saveLocked(module, tenants)
			}
		}
		return s.saveLocked(module, append(tenants, t))
	})
}

func (s *TenantStore) Delete(module, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return withFileLock(s.path(module), func() error {
		tenants, err := s.Load(module)
		if err != nil {
			return err
		}
		out := make([]domain.Tenant, 0, len(tenants))
		found := false
		for _, t := range tenants {
			if t.Name == name {
				found = true
				continue
			}
			out = append(out, t)
		}
		if !found {
			return domain.ErrTenantNotFound
		}
		return s.saveLocked(module, out)
	})
}
