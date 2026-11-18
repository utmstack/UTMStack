package upgrade

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

type v11Module struct {
	ID   int64
	Name string `gorm:"column:module_name"`
}

type v11ModuleGroup struct {
	ID        int64
	ModuleID  int64
	GroupName string
	GroupDesc string `gorm:"column:group_description"`
}

type v11ModuleGroupConf struct {
	GroupID int64
	Key     string `gorm:"column:conf_key"`
	Value   string `gorm:"column:conf_value"`
}

var v12PluginNames = map[string]string{
	"AWS_IAM_USER": "aws",
	"AZURE":        "azure",
	"GCP":          "gcp",
	"O365":         "o365",
	"BITDEFENDER":  "bitdefender",
	"CROWDSTRIKE":  "crowdstrike",
	"SOPHOS":       "sophos",
}

func v12PluginKey(moduleName string) string {
	if key, ok := v12PluginNames[strings.ToUpper(moduleName)]; ok {
		return key
	}
	return strings.ToLower(moduleName)
}

func convertPlugins(db *gorm.DB, yamlDir string) error {
	if !tableExists(db, "utm_module_group") {
		logSkipTable("utm_module_group")
		return nil
	}

	var modules []v11Module
	if err := db.Raw(`SELECT id, module_name FROM utm_module`).Scan(&modules).Error; err != nil {
		return err
	}
	moduleNames := map[int64]string{}
	for _, m := range modules {
		moduleNames[m.ID] = m.Name
	}

	var groups []v11ModuleGroup
	if err := db.Raw(`SELECT id, module_id, group_name, group_description FROM utm_module_group ORDER BY id`).Scan(&groups).Error; err != nil {
		return err
	}

	var confs []v11ModuleGroupConf
	if err := db.Raw(`SELECT group_id, conf_key, conf_value FROM utm_module_group_configuration`).Scan(&confs).Error; err != nil {
		return err
	}
	confByGroup := map[int64]map[string]string{}
	for _, c := range confs {
		if c.Key == "" {
			continue
		}
		if confByGroup[c.GroupID] == nil {
			confByGroup[c.GroupID] = map[string]string{}
		}
		confByGroup[c.GroupID][c.Key] = c.Value
	}

	groupsByPlugin := map[string][]pluginConfigGroup{}
	migrated := 0
	for _, g := range groups {
		moduleName := moduleNames[g.ModuleID]
		if moduleName == "" {
			catcher.Warn("v11 upgrade: plugin group with unknown module", map[string]any{"group_id": g.ID, "module_id": g.ModuleID})
			continue
		}
		if _, ok := v12PluginNames[strings.ToUpper(moduleName)]; !ok {
			continue
		}
		key := v12PluginKey(moduleName)
		config := map[string]string{}
		for k, v := range confByGroup[g.ID] {
			if v == "" {
				continue
			}
			config[k] = v
		}
		groupsByPlugin[key] = append(groupsByPlugin[key], pluginConfigGroup{
			Name:        g.GroupName,
			Description: g.GroupDesc,
			Config:      config,
		})
		migrated++
	}

	for key, gs := range groupsByPlugin {
		sort.Slice(gs, func(i, j int) bool { return gs[i].Name < gs[j].Name })
		if err := mergePluginGroups(filepath.Join(yamlDir, "system_plugins_"+key+".yaml"), key, gs); err != nil {
			return err
		}
	}

	catcher.Info("v11 upgrade: plugin groups converted", map[string]any{"groups": migrated})
	return nil
}

type pluginConfigGroup struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Config      map[string]string `yaml:"config"`
}

type pluginTenantConfig struct {
	ID     uuid.UUID           `yaml:"id"`
	Groups []pluginConfigGroup `yaml:"groups"`
}

type pluginEntry struct {
	Tenants []pluginTenantConfig `yaml:"tenants"`
}

type pluginsFile struct {
	Plugins map[string]pluginEntry `yaml:"plugins"`
}

func mergePluginGroups(path, key string, groups []pluginConfigGroup) error {
	pf := pluginsFile{Plugins: map[string]pluginEntry{}}
	if data, err := os.ReadFile(path); err == nil {
		if err := yaml.Unmarshal(data, &pf); err != nil {
			return err
		}
		if pf.Plugins == nil {
			pf.Plugins = map[string]pluginEntry{}
		}
	}

	entry := pf.Plugins[key]
	var existing *pluginTenantConfig
	for i := range entry.Tenants {
		if entry.Tenants[i].ID == defaultTenantID {
			existing = &entry.Tenants[i]
			break
		}
	}
	if existing == nil {
		entry.Tenants = append(entry.Tenants, pluginTenantConfig{ID: defaultTenantID})
		existing = &entry.Tenants[len(entry.Tenants)-1]
	}
	existing.Groups = groups
	pf.Plugins[key] = entry

	data, err := yaml.Marshal(pf)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
