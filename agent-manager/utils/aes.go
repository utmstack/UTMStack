package utils

import (
	"github.com/AtlasInsideCorp/AtlasInsideAES"
)

func DecryptValue(key string, value string) (string, error) {
	return AtlasInsideAES.AESDecrypt(value, []byte(key))
}
