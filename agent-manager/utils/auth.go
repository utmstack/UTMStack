package utils

import (
	"crypto/subtle"
)

type ConnectorAuth struct {
	Key      string
	TenantID string
}

func IsKeyPairValid(key string, id uint, cache map[uint]ConnectorAuth) (ConnectorAuth, bool) {
	entry, ok := cache[id]
	if !ok {
		return ConnectorAuth{}, false
	}
	if subtle.ConstantTimeCompare([]byte(key), []byte(entry.Key)) == 1 {
		return entry, true
	}
	return ConnectorAuth{}, false
}
