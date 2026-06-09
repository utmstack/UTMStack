package listeners

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	stdhttp "net/http"
	"os"
	"strings"
)

func validateAuth(inst *httpInstance, r *stdhttp.Request, body []byte) error {
	switch inst.auth {
	case "bearer":
		return validateBearer(r, inst.tokenPath)
	case "hmac":
		return validateHMAC(r, body, inst.tokenPath, inst.signatureHeader)
	default:
		return nil // no auth — open endpoint
	}
}

func validateBearer(r *stdhttp.Request, tokenPath string) error {
	token, err := readToken(tokenPath)
	if err != nil {
		return fmt.Errorf("load token: %w", err)
	}
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return fmt.Errorf("missing or invalid Authorization header")
	}
	provided := strings.TrimPrefix(authHeader, "Bearer ")
	if subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {
		return fmt.Errorf("invalid bearer token")
	}
	return nil
}

func validateHMAC(r *stdhttp.Request, body []byte, tokenPath, sigHeader string) error {
	token, err := readToken(tokenPath)
	if err != nil {
		return fmt.Errorf("load token: %w", err)
	}

	sig := r.Header.Get(sigHeader)
	if sig == "" {
		return fmt.Errorf("missing %s header", sigHeader)
	}

	rawSig := sig
	if idx := strings.Index(sig, "="); idx != -1 {
		rawSig = sig[idx+1:]
	}

	expected, err := hex.DecodeString(rawSig)
	if err != nil {
		return fmt.Errorf("invalid signature format in %s", sigHeader)
	}

	mac := hmac.New(sha256.New, []byte(token))
	mac.Write(body)
	computed := mac.Sum(nil)

	if !hmac.Equal(computed, expected) {
		return fmt.Errorf("HMAC signature mismatch")
	}
	return nil
}

func readToken(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("token file %s: %w", path, err)
	}
	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", fmt.Errorf("token file %s is empty", path)
	}
	return token, nil
}

func GenerateTokenFile(path string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	token := hex.EncodeToString(b)
	if err := os.WriteFile(path, []byte(token+"\n"), 0600); err != nil {
		return "", fmt.Errorf("write token file: %w", err)
	}
	return token, nil
}
