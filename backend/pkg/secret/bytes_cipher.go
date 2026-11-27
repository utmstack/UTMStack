package secret

import (
	"bytes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
)

// BytesCipher is the same scheme as Cipher (AES-128-CBC, key and IV from
// sha1+PBKDF2-SHA1, base64 output) but for binary payloads: the plaintext is
// base64-encoded first so any bytes round-trip, and Decrypt validates base64
// and padding instead of panicking on bad input.
type BytesCipher struct {
	blk cipher.Block
	iv  []byte
}

func NewBytesCipher(key string) (*BytesCipher, error) {
	blk, err := NewCipher(key)
	if err != nil {
		return nil, err
	}
	block, iv, err := blk.setKey()
	if err != nil {
		return nil, err
	}
	return &BytesCipher{blk: block, iv: iv}, nil
}

func (f *BytesCipher) Encrypt(plain []byte) ([]byte, error) {
	content := PKCS5Padding([]byte(base64.StdEncoding.EncodeToString(plain)), f.blk.BlockSize())
	ct := make([]byte, len(content))
	cipher.NewCBCEncrypter(f.blk, f.iv).CryptBlocks(ct, content)
	return []byte(base64.StdEncoding.EncodeToString(ct)), nil
}

func (f *BytesCipher) Decrypt(blob []byte) ([]byte, error) {
	// CBC panics on a non-block-aligned buffer, so validate shape first.
	raw, err := base64.StdEncoding.DecodeString(string(blob))
	if err != nil || len(raw) == 0 || len(raw)%f.blk.BlockSize() != 0 {
		return nil, errors.New("secret: invalid ciphertext")
	}
	plain := make([]byte, len(raw))
	cipher.NewCBCDecrypter(f.blk, f.iv).CryptBlocks(plain, raw)
	unpadded, err := pkcs5Unpad(plain)
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(string(unpadded))
}

func pkcs5Unpad(b []byte) ([]byte, error) {
	n := len(b)
	if n == 0 {
		return nil, errors.New("secret: empty padding")
	}
	pad := int(b[n-1])
	if pad == 0 || pad > n || pad > 16 {
		return nil, errors.New("secret: invalid padding")
	}
	if !bytes.Equal(bytes.Repeat([]byte{byte(pad)}, pad), b[n-pad:]) {
		return nil, errors.New("secret: invalid padding")
	}
	return b[:n-pad], nil
}
