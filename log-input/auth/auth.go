package auth

import (
	"context"
	"crypto/sha256"
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
	keyPrefixAPIKey    = "auth:apikey:"

	refreshLockPrefix = "auth:refreshing:"
	refreshLockTTL    = 10 * time.Second

	refreshWait     = 2 * time.Second
	refreshPollTick = 100 * time.Millisecond

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

func (s *Service) ConnectorValid(ctx context.Context, key string, id uint64, typ string) bool {
	if key == "" {
		return false
	}

	prefix, ok := connectorPrefix(typ)
	if !ok {
		return false
	}

	if stored, found := s.get(ctx, prefix+fmt.Sprint(id)); found {
		return stored == key
	}

	// A miss is not a rejection: the key set may not be cached yet.
	if !s.fillConnectors(ctx, typ) {
		return false
	}

	stored, found := s.get(ctx, prefix+fmt.Sprint(id))
	return found && stored == key
}

func (s *Service) APIKeyValid(ctx context.Context, apiKey, clientIP string) bool {
	if apiKey == "" {
		return false
	}

	cacheKey := keyPrefixAPIKey + hashed(apiKey+"\x00"+clientIP)
	if _, found := s.get(ctx, cacheKey); found {
		return true
	}

	if !s.be.authenticate(ctx, apiKey, clientIP) {
		return false
	}

	s.set(ctx, cacheKey, "1")
	return true
}

func (s *Service) InternalKeyValid(key string) bool {
	return s.cfg.InternalKey != "" && key == s.cfg.InternalKey
}

// fillConnectors resolves a miss by listing the connector set and caching all
// of it. The agent-manager has no lookup by id, so a miss costs a full listing
// and only the replica holding the lock pays for it.
func (s *Service) fillConnectors(ctx context.Context, typ string) bool {
	lock := refreshLockPrefix + typ

	cCtx, cancel := context.WithTimeout(ctx, redisTimeout)
	won, err := s.rdb.SetNX(cCtx, lock, "1", refreshLockTTL).Result()
	cancel()
	if err != nil {
		_ = catcher.Error("cannot take the auth refresh lock", err, map[string]any{
			"process": processName,
			"type":    typ,
		})
		return false
	}

	if !won {
		return s.waitForRefresh(ctx, lock)
	}
	defer func() {
		dCtx, dCancel := context.WithTimeout(context.WithoutCancel(ctx), redisTimeout)
		defer dCancel()
		_ = s.rdb.Del(dCtx, lock).Err()
	}()

	keys, err := s.am.listKeys(ctx, typ)
	if err != nil {
		_ = catcher.Error("cannot list connector keys", err, map[string]any{
			"process": processName,
			"type":    typ,
		})
		return false
	}

	prefix, ok := connectorPrefix(typ)
	if !ok {
		return false
	}

	pipe := s.rdb.Pipeline()
	for id, key := range keys {
		pipe.Set(ctx, prefix+fmt.Sprint(id), key, s.cfg.AuthTTL)
	}
	pCtx, pCancel := context.WithTimeout(ctx, redisTimeout)
	defer pCancel()
	if _, err := pipe.Exec(pCtx); err != nil {
		_ = catcher.Error("cannot cache connector keys", err, map[string]any{
			"process": processName,
			"type":    typ,
			"keys":    len(keys),
		})
		return false
	}

	return true
}

func (s *Service) waitForRefresh(ctx context.Context, lock string) bool {
	deadline := time.NewTimer(refreshWait)
	defer deadline.Stop()
	tick := time.NewTicker(refreshPollTick)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-deadline.C:
			return false
		case <-tick.C:
			cCtx, cancel := context.WithTimeout(ctx, redisTimeout)
			n, err := s.rdb.Exists(cCtx, lock).Result()
			cancel()
			if err == nil && n == 0 {
				return true
			}
		}
	}
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
