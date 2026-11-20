package coordination

import (
	"strings"
)

// SanitizeKVKey makes a caller-supplied cursor or lease key valid for a NATS
// KeyValue bucket. It replaces every byte outside the KV charset with '_',
// then strips any trailing dot the KV rules forbid.
func SanitizeKVKey(key string) string {
	var b strings.Builder
	b.Grow(len(key))
	for i := 0; i < len(key); i++ {
		c := key[i]
		if isKVKeyByte(c) {
			b.WriteByte(c)
		} else {
			b.WriteByte('_')
		}
	}
	return strings.TrimRight(b.String(), ".")
}

// SanitizeSubjectToken makes one path element of a job subject valid. The
// token must stay within [a-zA-Z0-9_-] so it carries no dot (which would
// split the subject), no wildcard, no whitespace and no unicode. An empty
// result is impossible: every byte is replaced, and the all-empty input
// yields a single '_'.
func SanitizeSubjectToken(token string) string {
	var b strings.Builder
	b.Grow(len(token))
	for i := 0; i < len(token); i++ {
		c := token[i]
		if isSubjectTokenByte(c) {
			b.WriteByte(c)
		} else {
			b.WriteByte('_')
		}
	}
	if b.Len() == 0 {
		return "_"
	}
	return b.String()
}

func isKVKeyByte(c byte) bool {
	switch {
	case c >= 'a' && c <= 'z':
		return true
	case c >= 'A' && c <= 'Z':
		return true
	case c >= '0' && c <= '9':
		return true
	}
	return c == '-' || c == '_' || c == '/' || c == '=' || c == '.'
}

func isSubjectTokenByte(c byte) bool {
	switch {
	case c >= 'a' && c <= 'z':
		return true
	case c >= 'A' && c <= 'Z':
		return true
	case c >= '0' && c <= '9':
		return true
	}
	return c == '-' || c == '_'
}
