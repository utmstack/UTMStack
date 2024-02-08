package util

import (
	"os"

	"github.com/AtlasInsideCorp/AtlasInsideAES"
	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/enum"
	"github.com/utmstack/UTMStack/agent-manager/models"
)

var (
	passphrase = os.Getenv(config.UTMEncryptionKeyEnv)
)

func EncryptDecryptConfValues(conf *models.AgentModuleConfiguration, action string) *models.AgentModuleConfiguration {

	if conf.ConfDatatype == enum.PasswordType {
		if action != "encrypt" {
			conf.ConfValue, _ = AtlasInsideAES.AESDecrypt(conf.ConfValue, []byte(passphrase))
		} else {
			conf.ConfValue, _ = AtlasInsideAES.AESEncrypt(conf.ConfValue, []byte(passphrase))
		}
	}
	return conf
}

func EncryptValue(value string) (string, error) {
	return AtlasInsideAES.AESEncrypt(value, []byte(passphrase))
}

func DecryptValue(value string) (string, error) {
	return AtlasInsideAES.AESDecrypt(value, []byte(passphrase))
}
