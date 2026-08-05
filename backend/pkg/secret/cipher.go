package secret

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	b64 "encoding/base64"
	"errors"

	"golang.org/x/crypto/pbkdf2"
)

type Cipher struct {
	key []byte
	iv  []byte
}

const (
	IterationCount = 65536
	SaltLength     = 16
	KeyLength      = 16
)

// NewCipher derives the AES key and IV from the input. Any non-empty
// string works — short keys are accepted because we hash them.
func NewCipher(key string) (*Cipher, error) {
	if key == "" {
		return nil, errors.New("empty encryption key")
	}
	return &Cipher{key: []byte(key)}, nil
}

func (self *Cipher) setKey() (cipher.Block, []byte, error) {
	h := sha1.New()
	h.Write(self.key)
	salt := h.Sum(nil)
	keyEnc := pbkdf2.Key(self.key, salt, IterationCount, KeyLength, sha1.New)
	block, err := aes.NewCipher(keyEnc)
	if err != nil {
		return nil, nil, err
	}
	return block, salt[:SaltLength], nil
}

func (self *Cipher) Encrypt(src string) (string, error) {
	if len(src) == 0 {
		return "", nil
	}
	blkEncrypt, ivEncrypt, err := self.setKey()
	if err != nil {
		return "", nil
	}
	ecb := cipher.NewCBCEncrypter(blkEncrypt, ivEncrypt)
	content := []byte(src)
	content = PKCS5Padding(content, blkEncrypt.BlockSize())
	crypted := make([]byte, len(content))
	ecb.CryptBlocks(crypted, content)
	base64 := b64.StdEncoding.EncodeToString(crypted)
	return base64, nil
}

func (self *Cipher) Decrypt(crypt string) (string, error) {
	encryptedData, _ := b64.StdEncoding.DecodeString(crypt)
	if len(crypt) == 0 {
		return "", nil
	}
	blk, iv, err := self.setKey()
	if err != nil {
		return "", nil
	}
	ecb := cipher.NewCBCDecrypter(blk, iv)
	decrypted := make([]byte, len(encryptedData))
	ecb.CryptBlocks(decrypted, encryptedData)
	return string(PKCS5Trimming(decrypted)), nil
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5Trimming(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}
