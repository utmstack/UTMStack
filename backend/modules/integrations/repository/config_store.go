package repository

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"

	"github.com/utmstack/utmstack/backend/modules/integrations/connectors"
	"github.com/utmstack/utmstack/backend/modules/integrations/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

func withFileLock(path string, fn func() error) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	lf, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	defer lf.Close()
	if err := syscall.Flock(int(lf.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	return fn()
}

var _ connectors.ConfigRepository = (*ConfigStore)(nil)

type ConfigStore struct {
	dir string
	mu  sync.Mutex
}

func NewConfigStore(dir string) *ConfigStore { return &ConfigStore{dir: dir} }

var pluginNames = map[string]string{
	"AWS_IAM_USER": "aws",
	"AZURE":        "azure",
	"GCP":          "gcp",
	"O365":         "o365",
	"BITDEFENDER":  "bitdefender",
	"CROWDSTRIKE":  "crowdstrike",
	"SOPHOS":       "sophos",
}

func pluginKey(integration string) string {
	name, ok := pluginNames[strings.ToUpper(integration)]
	if !ok {
		name = strings.ToLower(integration)
	}
	return name
}

func (s *ConfigStore) path(integration string) string {
	return filepath.Join(s.dir, "system_plugins_"+pluginKey(integration)+".yaml")
}

type pluginsFile struct {
	Plugins map[string]pluginEntry `yaml:"plugins"`
}

type pluginEntry struct {
	Tenants []domain.TenantConfig `yaml:"tenants"`
}

func callerTenant(ctx context.Context) (uuid.UUID, error) {
	raw := authz.TenantIDFromContext(ctx)
	if raw == "" {
		return uuid.Nil, tenancy.ErrNoTenant
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, tenancy.ErrNoTenant
	}
	return id, nil
}

func (s *ConfigStore) readFile(integration string) (pluginsFile, error) {
	pf := pluginsFile{Plugins: map[string]pluginEntry{}}

	data, err := os.ReadFile(s.path(integration))
	if errors.Is(err, os.ErrNotExist) {
		return pf, nil
	}
	if err != nil {
		return pf, err
	}
	if err := yaml.Unmarshal(data, &pf); err != nil {
		return pluginsFile{Plugins: map[string]pluginEntry{}}, err
	}
	if pf.Plugins == nil {
		pf.Plugins = map[string]pluginEntry{}
	}
	return pf, nil
}

func (s *ConfigStore) writeLocked(integration string, pf pluginsFile, tenants []domain.TenantConfig) error {
	if pf.Plugins == nil {
		pf.Plugins = map[string]pluginEntry{}
	}
	pf.Plugins[pluginKey(integration)] = pluginEntry{Tenants: tenants}

	data, err := yaml.Marshal(pf)
	if err != nil {
		return err
	}
	dst := s.path(integration)
	tmp := dst + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}

func (s *ConfigStore) Load(ctx context.Context, integration string) ([]domain.ConfigGroup, error) {
	tenant, err := callerTenant(ctx)
	if err != nil {
		return nil, err
	}

	pf, err := s.readFile(integration)
	if err != nil {
		return nil, err
	}
	for _, tc := range pf.Plugins[pluginKey(integration)].Tenants {
		if tc.ID == tenant {
			return tc.Groups, nil
		}
	}
	return []domain.ConfigGroup{}, nil
}

func (s *ConfigStore) LoadAllTenants(_ context.Context, integration string) ([]domain.TenantConfig, error) {
	pf, err := s.readFile(integration)
	if err != nil {
		return nil, err
	}
	return pf.Plugins[pluginKey(integration)].Tenants, nil
}

func (s *ConfigStore) Save(ctx context.Context, integration string, groups []domain.ConfigGroup) error {
	tenant, err := callerTenant(ctx)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return withFileLock(s.path(integration), func() error {
		pf, err := s.readFile(integration)
		if err != nil {
			return err
		}
		return s.writeLocked(integration, pf,
			replaceTenant(pf.Plugins[pluginKey(integration)].Tenants, tenant, groups))
	})
}

func (s *ConfigStore) Upsert(ctx context.Context, integration string, group domain.ConfigGroup) error {
	tenant, err := callerTenant(ctx)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return withFileLock(s.path(integration), func() error {
		pf, err := s.readFile(integration)
		if err != nil {
			return err
		}
		tenants := pf.Plugins[pluginKey(integration)].Tenants

		groups := groupsOf(tenants, tenant)
		replaced := false
		for i := range groups {
			if groups[i].Name == group.Name {
				groups[i] = group
				replaced = true
				break
			}
		}
		if !replaced {
			groups = append(groups, group)
		}
		return s.writeLocked(integration, pf, replaceTenant(tenants, tenant, groups))
	})
}

func (s *ConfigStore) Delete(ctx context.Context, integration, groupName string) error {
	tenant, err := callerTenant(ctx)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return withFileLock(s.path(integration), func() error {
		pf, err := s.readFile(integration)
		if err != nil {
			return err
		}
		tenants := pf.Plugins[pluginKey(integration)].Tenants

		groups := groupsOf(tenants, tenant)
		out := make([]domain.ConfigGroup, 0, len(groups))
		found := false
		for _, g := range groups {
			if g.Name == groupName {
				found = true
				continue
			}
			out = append(out, g)
		}
		if !found {
			return domain.ErrConfigGroupNotFound
		}
		return s.writeLocked(integration, pf, replaceTenant(tenants, tenant, out))
	})
}

func groupsOf(tenants []domain.TenantConfig, tenant uuid.UUID) []domain.ConfigGroup {
	for _, tc := range tenants {
		if tc.ID == tenant {
			return tc.Groups
		}
	}
	return nil
}

func replaceTenant(tenants []domain.TenantConfig, tenant uuid.UUID, groups []domain.ConfigGroup) []domain.TenantConfig {
	out := make([]domain.TenantConfig, 0, len(tenants)+1)
	seen := false
	for _, tc := range tenants {
		if tc.ID != tenant {
			out = append(out, tc)
			continue
		}
		seen = true
		if len(groups) > 0 {
			tc.Groups = groups
			out = append(out, tc)
		}
	}
	if !seen && len(groups) > 0 {
		out = append(out, domain.TenantConfig{ID: tenant, Groups: groups})
	}
	return out
}
