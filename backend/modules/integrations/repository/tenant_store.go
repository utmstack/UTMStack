package repository

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/utmstack/utmstack/backend/modules/integrations/connectors"
	"github.com/utmstack/utmstack/backend/modules/integrations/domain"
	"gopkg.in/yaml.v3"
)

const disabledSuffix = ".disabled"

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

func tenantFileName(module string) string {
	name, ok := pluginNames[strings.ToUpper(module)]
	if !ok {
		name = strings.ToLower(module)
	}
	return "system_plugins_" + name + ".yaml"
}

func (s *TenantStore) enabledPath(module string) string {
	return filepath.Join(s.dir, tenantFileName(module))
}

func (s *TenantStore) disabledPath(module string) string {
	return s.enabledPath(module) + disabledSuffix
}

func (s *TenantStore) path(module string) string {
	en := s.enabledPath(module)
	if _, err := os.Stat(en); err == nil {
		return en
	}
	if dis := s.disabledPath(module); fileExists(dis) {
		return dis
	}
	return en
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func (s *TenantStore) SetActiveByModule(_ context.Context, module string, active bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	en := s.enabledPath(module)
	dis := s.disabledPath(module)
	if active {
		if fileExists(dis) {
			return os.Rename(dis, en)
		}
		return nil
	}
	if fileExists(en) {
		return os.Rename(en, dis)
	}
	return nil
}

func (s *TenantStore) Load(module string) ([]domain.Tenant, error) {
	data, err := os.ReadFile(s.path(module))
	if errors.Is(err, os.ErrNotExist) {
		return []domain.Tenant{}, nil
	}
	if err != nil {
		return nil, err
	}
	var tenants []domain.Tenant
	if err := yaml.Unmarshal(data, &tenants); err != nil {
		return nil, err
	}
	return tenants, nil
}

func (s *TenantStore) Save(module string, tenants []domain.Tenant) error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(tenants)
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

	tenants, err := s.Load(module)
	if err != nil {
		return err
	}
	for i := range tenants {
		if tenants[i].Name == t.Name {
			tenants[i] = t
			return s.Save(module, tenants)
		}
	}
	return s.Save(module, append(tenants, t))
}

func (s *TenantStore) Delete(module, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

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
	return s.Save(module, out)
}
