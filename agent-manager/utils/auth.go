package utils

import (
	"crypto/subtle"
)

func IsKeyPairValid(key string, id uint, cache map[uint]string) (string, bool) {
	agentKey, ok := cache[id]
	if !ok {
		return "", false
	}
	if subtle.ConstantTimeCompare([]byte(key), []byte(agentKey)) == 1 {
		return agentKey, true
	}
	return "", false
}
