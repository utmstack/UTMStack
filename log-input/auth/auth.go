package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/UTMStack/log-input/config"
)

const (
	processName = "log-input"

	keyPrefixAgent     = "auth:agent:"
	keyPrefixCollector = "auth:collector:"
	keyPrefixAPIKey    = "auth:apikey:v2:"

	redisTimeout = 3 * time.Second
)

type Service struct {
	rdb *redis.Client
	cfg *config.Config

	am *agentManagerClient
	be *backendClient
}

func New(cfg *config.Config) *Service {
	return &Service{
		rdb: redis.NewClient(&redis.Options{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		}),
		cfg: cfg,
		am:  newAgentManagerClient(cfg),
		be:  newBackendClient(cfg),
	}
}

func (s *Service) Close() error { return s.rdb.Close() }

func (s *Service) Ping(ctx context.Context) error {
	cCtx, cancel := context.WithTimeout(ctx, redisTimeout)
	defer cancel()
	return s.rdb.Ping(cCtx).Err()
}

// ConnectorTenant authenticates a connector and reports which tenant it enrolled
// into. The tenant comes from the credential, never from what the pusher sends:
// otherwise anyone holding one key could write into any tenant.
func (s *Service) ConnectorTenant(ctx context.Context, key string, id uint64, typ string) (string, bool) {
	if key == "" {
		return "", false
	}

	prefix, ok := connectorPrefix(typ)
	if !ok {
		return "", false
	}

	if stored, found := s.get(ctx, prefix+fmt.Sprint(id)); found {
		if tenant, match := matchConnector(stored, key); match {
			return tenant, true
		}
		return "", false
	}

	auth, err := s.am.getKey(ctx, typ, id)
	if err != nil {
		_ = catcher.Error("cannot resolve a connector key", err, map[string]any{
			"process": processName,
			"type":    typ,
			"id":      id,
		})
		return "", false
	}

	stored := auth.TenantID + "\x00" + auth.Key
	s.set(ctx, prefix+fmt.Sprint(id), stored)

	return matchConnector(stored, key)
}

func matchConnector(stored, key string) (string, bool) {
	tenant, storedKey, ok := strings.Cut(stored, "\x00")
	if !ok {
		tenant, storedKey = "", stored
	}
	if subtle.ConstantTimeCompare([]byte(storedKey), []byte(key)) != 1 {
		return "", false
	}
	return tenant, true
}

func (s *Service) APIKeyTenant(ctx context.Context, apiKey, clientIP string) (string, bool) {
	if apiKey == "" {
		return "", false
	}

	cacheKey := keyPrefixAPIKey + hashed(apiKey+"\x00"+clientIP)
	if tenant, found := s.get(ctx, cacheKey); found {
		return tenant, true
	}

	tenant, ok := s.be.authenticate(ctx, apiKey, clientIP)
	if !ok {
		return "", false
	}

	s.set(ctx, cacheKey, tenant)
	return tenant, true
}

func (s *Service) InternalKeyValid(key string) bool {
	return s.cfg.InternalKey != "" && key == s.cfg.InternalKey
}

// get reads a cached answer. An unreachable Redis reads as a miss so it cannot
// authenticate anyone by accident.
func (s *Service) get(ctx context.Context, key string) (string, bool) {
	cCtx, cancel := context.WithTimeout(ctx, redisTimeout)
	defer cancel()

	v, err := s.rdb.Get(cCtx, key).Result()
	if err == redis.Nil {
		return "", false
	}
	if err != nil {
		_ = catcher.Error("cannot read the auth cache", err, map[string]any{
			"process": processName,
		})
		return "", false
	}
	return v, true
}

func (s *Service) set(ctx context.Context, key, value string) {
	cCtx, cancel := context.WithTimeout(ctx, redisTimeout)
	defer cancel()

	if err := s.rdb.Set(cCtx, key, value, s.cfg.AuthTTL).Err(); err != nil {
		_ = catcher.Error("cannot write the auth cache", err, map[string]any{
			"process": processName,
		})
	}
}

func connectorPrefix(typ string) (string, bool) {
	switch strings.ToLower(typ) {
	case "agent":
		return keyPrefixAgent, true
	case "collector":
		return keyPrefixCollector, true
	}
	return "", false
}

// hashed keeps the API key out of Redis: what is cached is that it was
// accepted, not the key.
func hashed(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
