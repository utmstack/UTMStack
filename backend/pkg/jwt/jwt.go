package jwt

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const Issuer = "utmstack-backend"

type Claims struct {
	Login       string   `json:"login"`
	Email       string   `json:"email,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	SessionID   uint64   `json:"sid,omitempty"`
	TenantID    string   `json:"tid,omitempty"`
	jwt.RegisteredClaims
}

func (c *Claims) UserID() (uint64, error) {
	if c.Subject == "" {
		return 0, errors.New("missing subject claim")
	}
	id, err := strconv.ParseUint(c.Subject, 10, 64)
	if err != nil {
		return 0, errors.New("invalid subject claim")
	}
	return id, nil
}

type Signer struct {
	secret []byte
	ttl    time.Duration
}

func NewSigner(secret string, ttl time.Duration) *Signer {
	return &Signer{secret: []byte(secret), ttl: ttl}
}

func (s *Signer) Sign(userID uint64, login, email string, permissions, roles []string, sessionID uint64, tenantID string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(s.ttl)
	claims := Claims{
		Login:       login,
		Email:       email,
		Permissions: permissions,
		Roles:       roles,
		SessionID:   sessionID,
		TenantID:    tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Issuer:    Issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

func (s *Signer) Verify(tokenStr string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
