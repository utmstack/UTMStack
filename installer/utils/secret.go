package utils

import (
	"crypto/rand"
	"math/big"
)

func GenerateSecret(size int) string {
	var characters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	if size <= 0 {
		return ""
	}

	result := make([]rune, size)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(characters))))
		if err != nil {
			panic(err) // Consider returning error if signature allows, but keeping panic for now as panic on CSPRNG fail is reasonable for secret gen
		}
		result[i] = characters[num.Int64()]
	}
	return string(result)
}
