package configuration

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	aesCrypt "github.com/AtlasInsideCorp/AtlasInsideAES"
	"github.com/utmstack/UTMStack/agent/agent/utils"
)

type Config struct {
	Server             string `yaml:"server"`
	AgentID            uint   `yaml:"agent-id"`
	AgentKey           string `yaml:"agent-key"`
	SkipCertValidation bool   `yaml:"insecure"`
}

func GetInitialConfig() (*Config, string) {
	cnf := Config{
		Server: os.Args[2],
	}
	skip := os.Args[4]
	if skip == "yes" {
		cnf.SkipCertValidation = true
	} else {
		cnf.SkipCertValidation = false
	}
	return &cnf, os.Args[3]
}

var (
	cnf      Config
	confOnce sync.Once
)

func GetCurrentConfig() (*Config, error) {
	cnf = Config{}
	var errR error
	confOnce.Do(func() {
		path, err := utils.GetMyPath()
		if err != nil {
			errR = fmt.Errorf("failed to get current path: %v", err)
			return
		}
		var encryptConfig Config
		if err = utils.ReadYAML(filepath.Join(path, "config.yml"), &encryptConfig); err != nil {
			errR = fmt.Errorf("error reading config file: %v", err)
			return
		}

		// Get key
		key, err := utils.GenerateKey(REPLACE_KEY)
		if err != nil {
			errR = fmt.Errorf("error geneating key: %v", err)
			return
		}

		// Decrypt config
		agentKey, err := aesCrypt.AESDecrypt(encryptConfig.AgentKey, key)
		if err != nil {
			errR = fmt.Errorf("error encoding agent key: %v", err)
			return
		}

		cnf.Server = encryptConfig.Server
		cnf.AgentID = encryptConfig.AgentID
		cnf.AgentKey = agentKey
		cnf.SkipCertValidation = encryptConfig.SkipCertValidation

	})
	if errR != nil {
		return nil, errR
	}
	return &cnf, nil
}

func SaveConfig(cnf *Config) error {
	// Get current path
	path, err := utils.GetMyPath()
	if err != nil {
		return fmt.Errorf("failed to get current path: %v", err)
	}

	// Get key
	key, err := utils.GenerateKey(REPLACE_KEY)
	if err != nil {
		return fmt.Errorf("error geneating key: %v", err)
	}

	// Encrypt config
	agentKey, err := aesCrypt.AESEncrypt(cnf.AgentKey, key)
	if err != nil {
		return fmt.Errorf("error encoding agent key: %v", err)
	}

	encryptConf := &Config{
		Server:             cnf.Server,
		AgentID:            cnf.AgentID,
		AgentKey:           agentKey,
		SkipCertValidation: cnf.SkipCertValidation,
	}

	// Write config in config.yml
	if err := utils.WriteYAML(filepath.Join(path, "config.yml"), encryptConf); err != nil {
		return err
	}
	return nil
}
