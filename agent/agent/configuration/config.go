package configuration

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	aesCrypt "github.com/AtlasInsideCorp/AtlasInsideAES"
	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/agent/agent/utils"
)

type InstallationUUID struct {
	UUID string `yaml:"uuid"`
}

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
	cnf          = Config{}
	confOnce     sync.Once
	instuuid     = ""
	instuuidOnce sync.Once
)

func GetCurrentConfig() (*Config, error) {
	var errR error
	confOnce.Do(func() {
		path, err := utils.GetMyPath()
		if err != nil {
			errR = fmt.Errorf("failed to get current path: %v", err)
			return
		}

		uuidExists := utils.CheckIfPathExist(filepath.Join(path, UUIDFileName))

		var encryptConfig Config
		if err = utils.ReadYAML(filepath.Join(path, "config.yml"), &encryptConfig); err != nil {
			errR = fmt.Errorf("error reading config file: %v", err)
			return
		}

		// Get key
		var key []byte
		if uuidExists {
			uuid, err := GetUUID()
			if err != nil {
				errR = fmt.Errorf("failed to get uuid: %v", err)
				return
			}

			key, err = utils.GenerateKeyByUUID(REPLACE_KEY, uuid)
			if err != nil {
				errR = fmt.Errorf("error geneating key: %v", err)
				return
			}
		} else {
			key, err = utils.GenerateKey(REPLACE_KEY)
			if err != nil {
				errR = fmt.Errorf("error geneating key: %v", err)
				return
			}
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

		if !uuidExists {
			if err := SaveConfig(&cnf); err != nil {
				errR = fmt.Errorf("error writing config file: %v", err)
				return
			}
		}
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

	uuid, err := GenerateNewUUID()
	if err != nil {
		return fmt.Errorf("failed to generate uuid: %v", err)
	}

	// Get key
	key, err := utils.GenerateKeyByUUID(REPLACE_KEY, uuid)
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

func GenerateNewUUID() (string, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate uuid: %v", err)
	}

	InstallationUUID := InstallationUUID{
		UUID: uuid.String(),
	}

	path, err := utils.GetMyPath()
	if err != nil {
		return "", fmt.Errorf("failed to get current path: %v", err)
	}

	if err = utils.WriteYAML(filepath.Join(path, UUIDFileName), InstallationUUID); err != nil {
		return "", fmt.Errorf("error writing uuid file: %v", err)
	}

	return InstallationUUID.UUID, nil
}

func GetUUID() (string, error) {
	var errR error
	instuuidOnce.Do(func() {
		path, err := utils.GetMyPath()
		if err != nil {
			errR = fmt.Errorf("failed to get current path: %v", err)
			return
		}

		var uuid = InstallationUUID{}
		if err = utils.ReadYAML(filepath.Join(path, UUIDFileName), &uuid); err != nil {
			errR = fmt.Errorf("error reading uuid file: %v", err)
			return
		}

		instuuid = uuid.UUID
	})

	if errR != nil {
		return "", errR
	}

	return instuuid, nil
}
