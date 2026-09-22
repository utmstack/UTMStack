package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

func DeriveTLSCertKey(collectorID string, collectorKey string) [32]byte {
	return sha256.Sum256([]byte("tls-certs:" + collectorID + ":" + collectorKey))
}

func OpenTLSCertEnvelope(key [32]byte, ciphertextB64, nonceB64 string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return nil, errors.New("tls cert envelope: invalid ciphertext encoding")
	}
	nonce, err := base64.StdEncoding.DecodeString(nonceB64)
	if err != nil {
		return nil, errors.New("tls cert envelope: invalid nonce encoding")
	}

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(nonce) != gcm.NonceSize() {
		return nil, errors.New("tls cert envelope: invalid nonce size")
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("tls cert envelope: decryption failed (wrong key or tampered payload)")
	}
	return plaintext, nil
}
