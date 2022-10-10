package utils

import (
	"math/rand"
	"time"
)

func GenerateSecret() string {
	var capitals = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	var lowers = []rune("abcdefghijklmnopqrstuvwxyz")
	var numbers = []rune("0123456789")

	var s string
	for {
		if len(s) >= 16 {
			break
		}
		rand.Seed(time.Now().UTC().UnixNano())
		slice := map[int]string{0: string(capitals[rand.Intn(len(capitals))]), 1: string(lowers[rand.Intn(len(lowers))]), 2: string(numbers[rand.Intn(len(numbers))]), 3: "."}
		for _, v := range(slice){
			s += v
		}
	}

	return s
}
