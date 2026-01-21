package utils

import (
	"github.com/AtlasInsideCorp/AtlasInsideAES"
)

func DecryptValue(encryptedValue string) (string, error) {
	passphrase := Getenv("ENCRYPTION_KEY")
	return AtlasInsideAES.AESDecrypt(encryptedValue, []byte(passphrase))
}
