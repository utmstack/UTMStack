// Package crypto provides symmetric encryption for payloads crossing NATS.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"

	"golang.org/x/crypto/pbkdf2"
)

const (
	iterationCount = 65536
	keyLength      = 16
)

// Derives both the AES-128-CBC key and its IV from one key string via PBKDF2.
// This derivation must not change: it is the already-deployed scheme, so any
// change makes existing ciphertext unreadable.
type Cipher struct {
	key []byte
}

func NewCipher(key string) *Cipher {
	return &Cipher{key: []byte(key)}
}

func (c *Cipher) setKey() (cipher.Block, []byte, error) {
	h := sha1.New()
	h.Write(c.key)
	salt := h.Sum(nil)
	keyEnc := pbkdf2.Key(c.key, salt, iterationCount, keyLength, sha1.New)
	block, err := aes.NewCipher(keyEnc)
	if err != nil {
		return nil, nil, err
	}
	return block, salt[:keyLength], nil
}

// Returns already-plaintext input unchanged rather than erroring, so values
// written before a key was configured stay readable.
func (c *Cipher) Decrypt(crypt string) (string, error) {
	if crypt == "" {
		return "", nil
	}
	encryptedData, err := base64.StdEncoding.DecodeString(crypt)
	if err != nil {
		return crypt, nil // not base64 -> already plaintext
	}
	blk, iv, err := c.setKey()
	if err != nil {
		return crypt, err
	}
	if len(encryptedData)%aes.BlockSize != 0 {
		return crypt, nil // not a valid CBC block -> already plaintext
	}
	dec := cipher.NewCBCDecrypter(blk, iv)
	decrypted := make([]byte, len(encryptedData))
	dec.CryptBlocks(decrypted, encryptedData)
	return string(pkcs5Trim(decrypted)), nil
}

func (c *Cipher) Encrypt(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	blk, iv, err := c.setKey()
	if err != nil {
		return "", err
	}
	padded := pkcs5Pad([]byte(plain), aes.BlockSize)
	enc := cipher.NewCBCEncrypter(blk, iv)
	encrypted := make([]byte, len(padded))
	enc.CryptBlocks(encrypted, padded)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func pkcs5Trim(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	padding := int(data[len(data)-1])
	if padding > len(data) || padding == 0 {
		return data
	}
	return data[:len(data)-padding]
}

func pkcs5Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padded := make([]byte, len(data)+padding)
	copy(padded, data)
	for i := len(data); i < len(padded); i++ {
		padded[i] = byte(padding)
	}
	return padded
}
