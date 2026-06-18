package jwt

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const PreAuthAudience = "utmstack-tfa-preauth"

type PreAuthClaims struct {
	Method string `json:"tfa_method"`
	jwt.RegisteredClaims
}

func (c *PreAuthClaims) UserID() (uint64, error) {
	if c.Subject == "" {
		return 0, errors.New("missing subject claim")
	}
	id, err := strconv.ParseUint(c.Subject, 10, 64)
	if err != nil {
		return 0, errors.New("invalid subject claim")
	}
	return id, nil
}

type PreAuthSigner struct {
	secret []byte
	ttl    time.Duration
}

func NewPreAuthSigner(secret string, ttl time.Duration) *PreAuthSigner {
	return &PreAuthSigner{secret: []byte(secret), ttl: ttl}
}

func (s *PreAuthSigner) Sign(userID uint64, method string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(s.ttl)
	claims := PreAuthClaims{
		Method: method,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Issuer:    Issuer,
			Audience:  jwt.ClaimStrings{PreAuthAudience},
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

func (s *PreAuthSigner) Verify(tokenStr string) (*PreAuthClaims, error) {
	parsed, err := jwt.ParseWithClaims(tokenStr, &PreAuthClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*PreAuthClaims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	hasAud := false
	for _, a := range claims.Audience {
		if a == PreAuthAudience {
			hasAud = true
			break
		}
	}
	if !hasAud {
		return nil, errors.New("invalid token audience")
	}
	return claims, nil
}
