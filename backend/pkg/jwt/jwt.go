package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const Issuer = "utmstack-backend"

type Claims struct {
	Email       string    `json:"email,omitempty"`
	Permissions []string  `json:"permissions,omitempty"`
	Roles       []string  `json:"roles,omitempty"`
	SessionID   uuid.UUID `json:"sid,omitempty"`
	TenantID    string    `json:"tid,omitempty"`
	// Platform marks the instance operator, so a browser can hide what only
	// they may reach. The server still decides; this is for the UI.
	Platform bool `json:"platform,omitempty"`
	jwt.RegisteredClaims
}

func (c *Claims) UserID() (uuid.UUID, error) {
	if c.Subject == "" {
		return uuid.Nil, errors.New("missing subject claim")
	}
	id, err := uuid.Parse(c.Subject)
	if err != nil {
		return uuid.Nil, errors.New("invalid subject claim")
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

func (s *Signer) Sign(userID uuid.UUID, email string, permissions, roles []string, sessionID uuid.UUID, tenantID string, platform bool) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(s.ttl)
	claims := Claims{
		Email:       email,
		Permissions: permissions,
		Roles:       roles,
		SessionID:   sessionID,
		TenantID:    tenantID,
		Platform:    platform,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
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
