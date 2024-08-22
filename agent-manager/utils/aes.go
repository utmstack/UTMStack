package utils

import (
	"github.com/AtlasInsideCorp/AtlasInsideAES"
	"github.com/utmstack/UTMStack/agent-manager/config"
)

func DecryptValue(value string) (string, error) {
	return AtlasInsideAES.AESDecrypt(value, []byte(config.GetEncryptionKey()))
}
