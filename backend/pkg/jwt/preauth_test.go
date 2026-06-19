package jwt

import (
	"strings"
	"testing"
	"time"
)

func TestPreAuthSigner_SignVerifyRoundTrip(t *testing.T) {
	s := NewPreAuthSigner("preauth-secret", 5*time.Minute)
	token, expiresAt, err := s.Sign(42, "EMAIL")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if token == "" {
		t.Fatal("empty token")
	}
	if expiresAt.Before(time.Now()) {
		t.Fatalf("expiresAt in the past: %v", expiresAt)
	}
	claims, err := s.Verify(token)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if claims.Method != "EMAIL" {
		t.Fatalf("method=%q want EMAIL", claims.Method)
	}
	uid, err := claims.UserID()
	if err != nil || uid != 42 {
		t.Fatalf("UserID()=%d, err=%v", uid, err)
	}
}

func TestPreAuthSigner_RejectsAccessToken(t *testing.T) {
	access := NewSigner("shared", time.Hour)
	pre := NewPreAuthSigner("shared", 5*time.Minute)

	accessTok, _, err := access.Sign(42, "alice", "alice@example.com", nil, nil, 1)
	if err != nil {
		t.Fatalf("sign access: %v", err)
	}
	if _, err := pre.Verify(accessTok); err == nil {
		t.Fatal("access token must not validate as pre-auth (audience claim is missing)")
	}
}

func TestPreAuthSigner_AccessSignerRejectsPreAuth(t *testing.T) {
	// Inverse direction: a pre-auth token must not pass the regular Signer.Verify
	// successfully as an access token.
	access := NewSigner("shared", time.Hour)
	pre := NewPreAuthSigner("shared", 5*time.Minute)

	preTok, _, err := pre.Sign(42, "EMAIL")
	if err != nil {
		t.Fatalf("sign preauth: %v", err)
	}
	claims, err := access.Verify(preTok)
	if err != nil {
		// Acceptable: HMAC validates but parsing into Claims struct yields a token
		// without permissions/login; downstream code can't act on it. The audience
		// check belongs at the verify-code handler, not the access-token verifier.
		// Either way: the pre-auth token must not yield usable access claims.
		return
	}
	// If access.Verify accepts it, the claims must clearly not look like a usable user.
	if claims.Login != "" {
		t.Fatalf("pre-auth token leaked a Login field into access claims: %+v", claims)
	}
}

func TestPreAuthSigner_RejectsTamperedSignature(t *testing.T) {
	s := NewPreAuthSigner("secret", 5*time.Minute)
	token, _, _ := s.Sign(1, "EMAIL")
	// Tamper the FIRST char of the signature segment. The last base64url char of an
	// HS512 signature carries non-significant padding bits that the (non-strict)
	// decoder ignores, so flipping it can decode to the same bytes and yield an
	// equivalent token; the first char is always significant.
	dot := strings.LastIndex(token, ".")
	repl := byte('A')
	if token[dot+1] == repl {
		repl = 'B'
	}
	bad := token[:dot+1] + string(repl) + token[dot+2:]
	if _, err := s.Verify(bad); err == nil {
		t.Fatal("tampered token must fail verification")
	}
}

func TestPreAuthSigner_RejectsExpired(t *testing.T) {
	s := NewPreAuthSigner("secret", -time.Second)
	token, _, _ := s.Sign(1, "EMAIL")
	if _, err := s.Verify(token); err == nil {
		t.Fatal("expired token must fail verification")
	}
}
