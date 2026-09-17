package crypto

import (
	"fmt"
	"strings"

	"github.com/AtlasInsideCorp/AtlasInsideAES"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/modules-config/config"
)

const (
	confTypePassword = "password"
	confTypeFile     = "file"
	moduleGCP        = "GCP"
)

func DecryptConfigurationSection(section *config.ConfigurationSection, key string) error {
	if section == nil {
		return nil
	}

	for _, group := range section.ModuleGroups {
		if err := decryptGroupConfigurations(section.ModuleName, group, key); err != nil {
			return err
		}
	}

	return nil
}

func DecryptModuleGroup(moduleName string, group *config.ModuleGroup, key string) error {
	if group == nil {
		return nil
	}

	return decryptGroupConfigurations(moduleName, group, key)
}

func decryptGroupConfigurations(moduleName string, group *config.ModuleGroup, key string) error {
	if group == nil {
		return nil
	}

	for _, cnf := range group.ModuleGroupConfigurations {
		if !shouldDecrypt(moduleName, cnf.ConfDataType, cnf.ConfValue) {
			continue
		}

		plain, err := safeAESDecrypt(cnf.ConfValue, key)
		if err != nil {
			return catcher.Error("failed to decrypt configuration value", err, map[string]any{
				"process":      "plugin_com.utmstack.modules-config",
				"module":       moduleName,
				"groupId":      group.Id,
				"confKey":      cnf.ConfKey,
				"confDataType": cnf.ConfDataType,
				"cipherLen":    len(cnf.ConfValue),
			})
		}

		cnf.ConfValue = plain
	}

	return nil
}

func safeAESDecrypt(cipherText, key string) (plain string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("decryption failed (malformed ciphertext or wrong key): %v", r)
		}
	}()
	return AtlasInsideAES.AESDecrypt(cipherText, []byte(key))
}

func shouldDecrypt(moduleName, confDataType, confValue string) bool {
	if confValue == "" {
		return false
	}

	dataType := strings.ToLower(strings.TrimSpace(confDataType))
	switch dataType {
	case confTypePassword:
		return true
	case confTypeFile:
		return strings.EqualFold(moduleName, moduleGCP)
	default:
		return false
	}
}
