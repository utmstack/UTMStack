package server

import (
	"testing"

	"github.com/utmstack/UTMStack/plugins/bitdefender/utils"
)

func TestAuthMatches(t *testing.T) {
	const key = "test-connection-key-0123456789"
	valid := utils.GenerateAuthCode(key)
	// Same length as valid, but derived from a different key.
	wrongKeySameLen := utils.GenerateAuthCode("other-connection-key-0123456789")

	tests := []struct {
		name string
		key  string
		hdr  string
		want bool
	}{
		{name: "valid header authorizes", key: key, hdr: valid, want: true},
		{name: "wrong key same length rejects", key: key, hdr: wrongKeySameLen, want: false},
		{name: "truncated header rejects", key: key, hdr: valid[:len(valid)-1], want: false},
		{name: "empty header rejects", key: key, hdr: "", want: false},
		{name: "empty key rejects", key: "", hdr: valid, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := authMatches(tt.key, tt.hdr); got != tt.want {
				t.Fatalf("authMatches(%q, %q) = %v, want %v", tt.key, tt.hdr, got, tt.want)
			}
		})
	}
}
