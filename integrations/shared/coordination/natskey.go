package coordination

import "encoding/base64"

// NatsToken encodes an arbitrary caller-supplied name for use as a NATS KV key
// or subject token. base64url (no padding) is injective and its alphabet stays
// within both charsets, so distinct names never share a key.
func NatsToken(name string) string {
	if name == "" {
		// NATS subjects and KV keys cannot carry an empty token.
		return "_"
	}
	return base64.RawURLEncoding.EncodeToString([]byte(name))
}
