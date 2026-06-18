package config

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/collectors/forwarder/utils"
	"github.com/utmstack/UTMStack/shared/fs"
)

type InstallationUUID struct {
	UUID string `yaml:"uuid"`
}

type Config struct {
	Server             string `yaml:"server"`
	CollectorID  uint   `yaml:"collector-id"`
	CollectorKey string `yaml:"collector-key"`
	SkipCertValidation bool   `yaml:"insecure"`
}

var (
	cnf                = Config{}
	confOnce           sync.Once
	installationId     = ""
	installationIdOnce sync.Once
)

func GetCurrentConfig() (*Config, error) {
	var errR error
	confOnce.Do(func() {
		var encryptConfig Config
		if err := fs.ReadYAML(ConfigurationFile, &encryptConfig); err != nil {
			errR = fmt.Errorf("error reading config file: %v", err)
			return
		}

		id, err := GetUUID()
		if err != nil {
			errR = fmt.Errorf("failed to get uuid: %v", err)
			return
		}

		agentKey, err := utils.DecryptAES(encryptConfig.CollectorKey, REPLACE_KEY, id)
		if err != nil {
			errR = fmt.Errorf("error decrypting agent key: %v", err)
			return
		}

		cnf.Server = encryptConfig.Server
		cnf.CollectorID = encryptConfig.CollectorID
		cnf.CollectorKey = agentKey
		cnf.SkipCertValidation = encryptConfig.SkipCertValidation
	})
	if errR != nil {
		return nil, errR
	}
	return &cnf, nil
}

func SaveConfig(cnf *Config) error {
	id, err := GenerateNewUUID()
	if err != nil {
		return fmt.Errorf("failed to generate uuid: %v", err)
	}

	agentKey, err := utils.EncryptAES(cnf.CollectorKey, REPLACE_KEY, id)
	if err != nil {
		return fmt.Errorf("error encrypting agent key: %v", err)
	}

	encryptConf := &Config{
		Server:             cnf.Server,
		CollectorID:  cnf.CollectorID,
		CollectorKey: agentKey,
		SkipCertValidation: cnf.SkipCertValidation,
	}

	if err := fs.WriteYAML(ConfigurationFile, encryptConf); err != nil {
		return err
	}
	return nil
}

func GenerateNewUUID() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate uuid: %v", err)
	}

	installUUID := InstallationUUID{
		UUID: id.String(),
	}

	if err = fs.WriteYAML(UUIDFileName, installUUID); err != nil {
		return "", fmt.Errorf("error writing uuid file: %v", err)
	}

	return installUUID.UUID, nil
}

func GetUUID() (string, error) {
	var errR error
	installationIdOnce.Do(func() {
		var id = InstallationUUID{}
		if err := fs.ReadYAML(UUIDFileName, &id); err != nil {
			errR = fmt.Errorf("error reading uuid file: %v", err)
			return
		}

		installationId = id.UUID
	})

	if errR != nil {
		return "", errR
	}

	return installationId, nil
}
