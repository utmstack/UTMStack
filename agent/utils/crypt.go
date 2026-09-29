package utils

import (
	"encoding/base64"
	"fmt"

	aesCrypt "github.com/AtlasInsideCorp/AtlasInsideAES"
)

// GenerateKeyByUUID generates an AES key from baseKey and uuid.
func GenerateKeyByUUID(baseKey string, uuid string) ([]byte, error) {
	if baseKey == "" {
		return nil, fmt.Errorf("build secret is not set: this binary was built without the REPLACE_KEY ldflag, so the at-rest key would carry no secret")
	}

	data := []byte(baseKey + uuid)
	base64Key := base64.StdEncoding.EncodeToString(data)
	return []byte(base64Key), nil
}

// EncryptAES encrypts plaintext using AES with a key derived from baseKey and uuid.
func EncryptAES(plaintext, baseKey, uuid string) (string, error) {
	key, err := GenerateKeyByUUID(baseKey, uuid)
	if err != nil {
		return "", fmt.Errorf("error generating key: %v", err)
	}
	encrypted, err := aesCrypt.AESEncrypt(plaintext, key)
	if err != nil {
		return "", fmt.Errorf("error encrypting: %v", err)
	}
	return encrypted, nil
}

// DecryptAES decrypts ciphertext using AES with a key derived from baseKey and uuid.
func DecryptAES(ciphertext, baseKey, uuid string) (string, error) {
	key, err := GenerateKeyByUUID(baseKey, uuid)
	if err != nil {
		return "", fmt.Errorf("error generating key: %v", err)
	}
	decrypted, err := aesCrypt.AESDecrypt(ciphertext, key)
	if err != nil {
		return "", fmt.Errorf("error decrypting: %v", err)
	}
	return decrypted, nil
}
