package utils

import (
	"crypto/rand"
	"math/big"
	"slices"
)

func GenerateSecret(size int) string {
	characters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	return randomFromSet(characters, size)
}

// GenerateComplexSecret returns a random password that always contains at least
// one uppercase letter, one lowercase letter, one digit and one special
// character, so it passes strict password validators like OpenSearch's.
func GenerateComplexSecret(size int) string {
	upper := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	lower := []rune("abcdefghijklmnopqrstuvwxyz")
	digits := []rune("0123456789")
	special := []rune("!@#%^*_-+=")

	if size < 4 {
		size = 4
	}

	all := slices.Concat(upper, lower, digits, special)

	result := make([]rune, size)
	// Guarantee one character from each required class.
	result[0] = upper[randIndex(len(upper))]
	result[1] = lower[randIndex(len(lower))]
	result[2] = digits[randIndex(len(digits))]
	result[3] = special[randIndex(len(special))]
	for i := 4; i < size; i++ {
		result[i] = all[randIndex(len(all))]
	}

	// Shuffle so the guaranteed characters are not always at the front.
	for i := len(result) - 1; i > 0; i-- {
		j := randIndex(i + 1)
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

func randomFromSet(characters []rune, size int) string {
	result := make([]rune, size)
	for i := range size {
		result[i] = characters[randIndex(len(characters))]
	}
	return string(result)
}

func randIndex(n int) int {
	idx, _ := rand.Int(rand.Reader, big.NewInt(int64(n)))
	return int(idx.Int64())
}
