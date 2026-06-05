package connectors

import "github.com/utmstack/utmstack/backend/pkg/secret"

// VariableCipher encrypts/decrypts secret incident-variable values at rest.
type VariableCipher interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

type SecretCipherAdapter struct {
	c *secret.Cipher
}

func NewSecretCipherAdapter(c *secret.Cipher) VariableCipher {
	return &SecretCipherAdapter{c: c}
}

func (a *SecretCipherAdapter) Encrypt(plaintext string) (string, error) {
	return a.c.Encrypt(plaintext)
}

func (a *SecretCipherAdapter) Decrypt(ciphertext string) (string, error) {
	return a.c.Decrypt(ciphertext)
}
