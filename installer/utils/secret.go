package utils

import (
	"math/rand"
)

func GenerateSecret(size int) string {
	var characters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

	var s string
	for {
		if len(s) >= size {
			break
		}

		s += string(characters[rand.Intn(len(characters))])
	}

	return s
}
