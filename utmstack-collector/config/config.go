package config

import (
	"os"
	"sync"

	aesCrypt "github.com/AtlasInsideCorp/AtlasInsideAES"
	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/utmstack-collector/utils"
)

type MSGDS struct {
	DataSource string
	Message    string
}

type InstallationUUID struct {
	UUID string `yaml:"uuid"`
}

type Config struct {
	Server             string `yaml:"server"`
	CollectorID        uint   `yaml:"collector-id"`
	CollectorKey       string `yaml:"collector-key"`
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
	cnf                = Config{}
	confOnce           sync.Once
	installationId     = ""
	installationIdOnce sync.Once
)

func GetCurrentConfig() (*Config, error) {
	var errR error
	confOnce.Do(func() {
		uuidExists := utils.CheckIfPathExist(UUIDFileName)

		var encryptConfig Config
		if err := utils.ReadYAML(ConfigurationFile, &encryptConfig); err != nil {
			errR = utils.Logger.ErrorF("error reading config file: %v", err)
			return
		}

		var key []byte
		var err error
		if uuidExists {
			id, err := GetUUID()
			if err != nil {
				errR = utils.Logger.ErrorF("failed to get uuid: %v", err)
				return
			}

			key, err = utils.GenerateKeyByUUID(REPLACE_KEY, id)
			if err != nil {
				errR = utils.Logger.ErrorF("error geneating key: %v", err)
				return
			}
		} else {
			key, err = utils.GenerateKey(REPLACE_KEY)
			if err != nil {
				errR = utils.Logger.ErrorF("error geneating key: %v", err)
				return
			}
		}

		collectorKey, err := aesCrypt.AESDecrypt(encryptConfig.CollectorKey, key)
		if err != nil {
			errR = utils.Logger.ErrorF("error encoding collector key: %v", err)
			return
		}

		cnf.Server = encryptConfig.Server
		cnf.CollectorID = encryptConfig.CollectorID
		cnf.CollectorKey = collectorKey
		cnf.SkipCertValidation = encryptConfig.SkipCertValidation

		if !uuidExists {
			if err := SaveConfig(&cnf); err != nil {
				errR = utils.Logger.ErrorF("error writing config file: %v", err)
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
	id, err := GenerateNewUUID()
	if err != nil {
		return utils.Logger.ErrorF("failed to generate uuid: %v", err)
	}

	key, err := utils.GenerateKeyByUUID(REPLACE_KEY, id)
	if err != nil {
		return utils.Logger.ErrorF("error geneating key: %v", err)
	}

	collectorKey, err := aesCrypt.AESEncrypt(cnf.CollectorKey, key)
	if err != nil {
		return utils.Logger.ErrorF("error encoding agent key: %v", err)
	}

	encryptConf := &Config{
		Server:             cnf.Server,
		CollectorID:        cnf.CollectorID,
		CollectorKey:       collectorKey,
		SkipCertValidation: cnf.SkipCertValidation,
	}

	if err := utils.WriteYAML(ConfigurationFile, encryptConf); err != nil {
		return err
	}
	return nil
}

func GenerateNewUUID() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", utils.Logger.ErrorF("failed to generate uuid: %v", err)
	}

	InstallationUUID := InstallationUUID{
		UUID: id.String(),
	}

	if err = utils.WriteYAML(UUIDFileName, InstallationUUID); err != nil {
		return "", utils.Logger.ErrorF("error writing uuid file: %v", err)
	}

	return InstallationUUID.UUID, nil
}

func GetUUID() (string, error) {
	var errR error
	installationIdOnce.Do(func() {
		var id = InstallationUUID{}
		if err := utils.ReadYAML(UUIDFileName, &id); err != nil {
			errR = utils.Logger.ErrorF("error reading uuid file: %v", err)
			return
		}

		installationId = id.UUID
	})

	if errR != nil {
		return "", errR
	}

	return installationId, nil
}
