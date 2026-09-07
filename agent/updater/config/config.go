package config

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/utmstack/UTMStack/shared/fs"
)

type Config struct {
	Server             string `json:"server" yaml:"server"`
	SkipCertValidation bool   `json:"insecure" yaml:"insecure"`
	// PauseAutoUpdate mirrors the same field in the agent's own config —
	// see C4 in GAPS_AND_IMPROVEMENTS.md. Both structs read the same
	// config.yml independently, each parsing the subset of fields it needs.
	PauseAutoUpdate bool `json:"pause-auto-update" yaml:"pause-auto-update"`
}

var (
	cnf      = Config{}
	confOnce sync.Once
)

func GetCurrentConfig() (*Config, error) {
	var errR error
	confOnce.Do(func() {
		configPath := filepath.Join(fs.GetExecutablePath(), "config.yml")

		var loadedConfig Config
		if err := fs.ReadYAML(configPath, &loadedConfig); err != nil {
			errR = fmt.Errorf("error reading config file: %v", err)
			return
		}

		cnf.Server = loadedConfig.Server
		cnf.SkipCertValidation = loadedConfig.SkipCertValidation
		cnf.PauseAutoUpdate = loadedConfig.PauseAutoUpdate
	})

	if errR != nil {
		return nil, errR
	}
	return &cnf, nil
}
