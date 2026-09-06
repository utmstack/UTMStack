package coordination

import (
	"encoding/json"

	"github.com/utmstack/UTMStack/plugins/shared/crypto"
)

// NATS has neither auth nor TLS here. Jobs stay unencrypted: their subject
// already carries tenant and group for routing.
func MarshalCursorPayload(v any, encryptionKey string) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if encryptionKey == "" {
		return data, nil
	}

	encrypted, err := crypto.NewCipher(encryptionKey).Encrypt(string(data))
	if err != nil {
		return nil, err
	}
	return []byte(encrypted), nil
}

// Cursors written before a key was configured stay readable: Decrypt passes
// already-plaintext input through unchanged.
func UnmarshalCursorPayload[T any](data []byte, encryptionKey string) (T, error) {
	var zero T

	raw := data
	if encryptionKey != "" {
		decrypted, err := crypto.NewCipher(encryptionKey).Decrypt(string(data))
		if err != nil {
			return zero, err
		}
		raw = []byte(decrypted)
	}

	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return zero, err
	}
	return out, nil
}
