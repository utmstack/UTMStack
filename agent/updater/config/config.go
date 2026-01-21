package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
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
		ex, err := os.Executable()
		if err != nil {
			errR = fmt.Errorf("error getting executable path: %v", err)
			return
		}
		exPath := filepath.Dir(ex)

		configPath := filepath.Join(exPath, "config.yml")
		content, err := os.ReadFile(configPath)
		if err != nil {
			errR = fmt.Errorf("error reading config file: %v", err)
			return
		}

		var loadedConfig Config
		err = yaml.Unmarshal(content, &loadedConfig)
		if err != nil {
			errR = fmt.Errorf("error parsing config file: %v", err)
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
