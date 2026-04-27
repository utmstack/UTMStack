package crypto

import (
	"fmt"
	"runtime/debug"
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
		decryptGroupConfigurations(section.ModuleName, group, key)
	}

	return nil
}

func DecryptModuleGroup(moduleName string, group *config.ModuleGroup, key string) error {
	if group == nil {
		return nil
	}

	decryptGroupConfigurations(moduleName, group, key)
	return nil
}

func decryptGroupConfigurations(moduleName string, group *config.ModuleGroup, key string) {
	if group == nil {
		return
	}
	// Remove after testing, before release to production
	for _, cnf := range group.ModuleGroupConfigurations {
		catcher.Info("crypto: evaluating field", map[string]any{
			"process":      "plugin_com.utmstack.modules-config",
			"module":       moduleName,
			"groupId":      group.Id,
			"confKey":      cnf.ConfKey,
			"confDataType": cnf.ConfDataType,
			"valueLen":     len(cnf.ConfValue),
			"confValue":    cnf.ConfValue,
			"key":          key,
			"keyLen":       len(key),
		})
		// Remove after testing, before release to production
		if !shouldDecrypt(moduleName, cnf.ConfDataType, cnf.ConfValue) {
			catcher.Info("crypto: skipped (shouldDecrypt=false)", map[string]any{
				"process":      "plugin_com.utmstack.modules-config",
				"module":       moduleName,
				"confKey":      cnf.ConfKey,
				"confDataType": cnf.ConfDataType,
			})
			continue
		}
		// Remove after testing, before release to production
		plain, err := safeAESDecrypt(cnf.ConfValue, key)
		if err != nil {
			_ = catcher.Error("failed to decrypt configuration value", err, map[string]any{
				"process":      "plugin_com.utmstack.modules-config",
				"module":       moduleName,
				"groupId":      group.Id,
				"confKey":      cnf.ConfKey,
				"confDataType": cnf.ConfDataType,
				"valueLen":     len(cnf.ConfValue),
				"confValue":    cnf.ConfValue,
				"key":          key,
				"keyLen":       len(key),
			})
			continue
		}
		// Remove after testing, before release to production
		catcher.Info("crypto: decrypted OK", map[string]any{
			"process":   "plugin_com.utmstack.modules-config",
			"module":    moduleName,
			"groupId":   group.Id,
			"confKey":   cnf.ConfKey,
			"plainLen":  len(plain),
			"plainHead": firstChars(plain, 64),
		})
		cnf.ConfValue = plain
	}
}

func safeAESDecrypt(cipherText, key string) (plain string, err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			err = fmt.Errorf("decryption panic recovered: %v | stack: %s", r, stack)
		}
	}()
	return AtlasInsideAES.AESDecrypt(cipherText, []byte(key))
}

func firstChars(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
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
