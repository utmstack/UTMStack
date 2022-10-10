package utils

import (
	"math/rand"
	"time"
)

func GenerateSecret(size int) string {
	var letters = []rune(".abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	rand.Seed(time.Now().UTC().UnixNano())

	dot := rand.Intn(size)

	s := make([]rune, size)
	for i := range s {
		if i != dot {
			s[i] = letters[rand.Intn(len(letters))]
		} else {
			s[i] = letters[0]
		}

	}

	return string(s)
}
