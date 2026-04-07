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
	})

	if errR != nil {
		return nil, errR
	}
	return &cnf, nil
}
