package collector

import (
	"context"
	"path/filepath"
	"time"

	aesCrypt "github.com/AtlasInsideCorp/AtlasInsideAES"
	"github.com/fsnotify/fsnotify"
	"github.com/utmstack/UTMStack/collectors/as400/config"
	"github.com/utmstack/UTMStack/collectors/as400/utils"
)

type AS400CollectorConfig struct {
	Servers []AS400ServerConfig `json:"servers" yaml:"servers"`
}

type AS400ServerConfig struct {
	Tenant   string `json:"tenant" yaml:"tenant"`
	Hostname string `json:"hostname" yaml:"hostname"`
	UserId   string `json:"userId" yaml:"userId"`
	Password string `json:"password" yaml:"password"`
}

// LocalConfigWatcher reads the admin-editable AS400 servers file and reloads on
// change (fsnotify), invoking onConfigChange with each new validated config.
// Configuration is local: there is no backend round-trip.
type LocalConfigWatcher struct {
	path           string
	onConfigChange func(*AS400CollectorConfig)
	current        *AS400CollectorConfig
}

func NewLocalConfigWatcher(path string, onConfigChange func(*AS400CollectorConfig)) *LocalConfigWatcher {
	return &LocalConfigWatcher{path: path, onConfigChange: onConfigChange}
}

func (w *LocalConfigWatcher) Start(ctx context.Context) {
	w.reload()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		utils.Logger.ErrorF("config watcher unavailable, changes will require a restart: %v", err)
		return
	}
	defer func() { _ = watcher.Close() }()

	// Watch the parent dir: editors commonly replace the file (rename/atomic
	// write), which a file-level watch would miss.
	if err := watcher.Add(filepath.Dir(w.path)); err != nil {
		utils.Logger.ErrorF("failed to watch config dir: %v", err)
		return
	}

	var debounce *time.Timer
	var debounceC <-chan time.Time
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-watcher.Events:
			if !ok {
				return
			}
			if filepath.Clean(ev.Name) != filepath.Clean(w.path) {
				continue
			}
			if debounce == nil {
				debounce = time.NewTimer(500 * time.Millisecond)
				debounceC = debounce.C
			} else {
				debounce.Reset(500 * time.Millisecond)
			}
		case <-debounceC:
			debounce = nil
			debounceC = nil
			w.reload()
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			utils.Logger.ErrorF("config watcher error: %v", err)
		}
	}
}

func (w *LocalConfigWatcher) reload() {
	cfg, err := loadLocalConfig(w.path)
	if err != nil {
		utils.Logger.ErrorF("failed to load config %s: %v", w.path, err)
		return
	}
	if err := validateConfig(cfg); err != nil {
		utils.Logger.ErrorF("invalid config: %v", err)
		return
	}
	if w.current != nil && configEquals(w.current, cfg) {
		return
	}
	w.current = cfg
	utils.Logger.Info("Config loaded: %d servers", len(cfg.Servers))
	if w.onConfigChange != nil {
		w.onConfigChange(cfg)
	}
}

// loadLocalConfig reads the manual YAML config. A missing file is treated as an
// empty config (no servers) so the collector idles until one is provided.
func loadLocalConfig(path string) (*AS400CollectorConfig, error) {
	cfg := &AS400CollectorConfig{Servers: []AS400ServerConfig{}}
	if !utils.CheckIfPathExist(path) {
		return cfg, nil
	}
	if err := utils.ReadYAML(path, cfg); err != nil {
		return nil, err
	}
	if cfg.Servers == nil {
		cfg.Servers = []AS400ServerConfig{}
	}
	return cfg, nil
}

func validateConfig(cfg *AS400CollectorConfig) error {
	for i, s := range cfg.Servers {
		if s.Tenant == "" || s.Hostname == "" || s.UserId == "" || s.Password == "" {
			return utils.Logger.ErrorF("server %d (%s): missing required fields (tenant, hostname, userId, password)", i, s.Tenant)
		}
	}
	return nil
}

func configEquals(a, b *AS400CollectorConfig) bool {
	if len(a.Servers) != len(b.Servers) {
		return false
	}
	for i := range a.Servers {
		if a.Servers[i].Tenant != b.Servers[i].Tenant ||
			a.Servers[i].Hostname != b.Servers[i].Hostname ||
			a.Servers[i].UserId != b.Servers[i].UserId ||
			a.Servers[i].Password != b.Servers[i].Password {
			return false
		}
	}
	return true
}

// encryptedCopy returns a copy of the config with passwords AES-encrypted, for
// the JAR config file. The input (plaintext) is left untouched so the watcher's
// equality check keeps comparing plaintext.
func encryptedCopy(cfg *AS400CollectorConfig) (*AS400CollectorConfig, error) {
	out := &AS400CollectorConfig{Servers: make([]AS400ServerConfig, len(cfg.Servers))}
	for i, s := range cfg.Servers {
		enc, err := aesCrypt.AESEncrypt(s.Password, []byte(config.REPLACE_KEY))
		if err != nil {
			return nil, err
		}
		s.Password = enc
		out.Servers[i] = s
	}
	return out, nil
}
