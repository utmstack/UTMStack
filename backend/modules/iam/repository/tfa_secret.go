package repository

import (
	"strings"

	"github.com/utmstack/utmstack/backend/pkg/secret"
)

const tfaEncPrefix = "enc:v1:"

func encryptTfaSecret(c *secret.Cipher, plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	crypt, err := c.Encrypt(plain)
	if err != nil {
		return "", err
	}
	return tfaEncPrefix + crypt, nil
}

func decryptTfaSecret(c *secret.Cipher, stored string) (string, error) {
	rest, tagged := strings.CutPrefix(stored, tfaEncPrefix)
	if !tagged {
		return stored, nil
	}
	return c.Decrypt(rest)
}
