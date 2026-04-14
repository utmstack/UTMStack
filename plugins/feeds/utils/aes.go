package utils

import (
	"github.com/AtlasInsideCorp/AtlasInsideAES"
	"github.com/threatwinds/go-sdk/plugins"
)

func DecryptValue(encryptedValue string) (string, error) {
	passphrase := plugins.PluginCfg("com.utmstack").Get("internalKey").String()
	return AtlasInsideAES.AESDecrypt(encryptedValue, []byte(passphrase))
}
