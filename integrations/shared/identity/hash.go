// Package identity derives a deterministic event identity from source fields,
// making a re-read event recognisable. Nothing collapses it today: utmstack.logs
// is a MergeTree whose ORDER BY excludes id, and alert dedup is per-rule.
package identity

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// Hash returns hex(sha256(fields joined by "|")). Field content is not escaped,
// and callers must include a tenant/group key so tenants cannot collide.
func Hash(fields ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(fields, "|")))
	return hex.EncodeToString(sum[:])
}
