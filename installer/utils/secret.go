package utils

import (
	"crypto/rand"
	"math/big"
)

func GenerateSecret(size int) string {
	characters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	charLen := big.NewInt(int64(len(characters)))

	result := make([]rune, size)
	for i := 0; i < size; i++ {
		idx, _ := rand.Int(rand.Reader, charLen)
		result[i] = characters[idx.Int64()]
	}

	return string(result)
}
