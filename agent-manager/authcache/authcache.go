package authcache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/threatwinds/go-sdk/catcher"
)

const (
	processName = "agent-manager"

	prefixAgent     = "auth:agent:"
	prefixCollector = "auth:collector:"

	// Long enough that a republish always renews it first, short enough that a
	// key nobody renews stops being trusted.
	keyTTL = 30 * time.Minute

	// Heals whatever the event-driven path missed: an eviction, a write that
	// failed, a Redis that was restarted.
	republishEvery = 5 * time.Minute

	opTimeout = 3 * time.Second
)

type Publisher struct {
	rdb *redis.Client
}

// New returns nil when no Redis is configured, and every method on a nil
// Publisher is a no-op — an install without Redis keeps working.
func New(addr, password string, db int) *Publisher {
	if addr == "" {
		return nil
	}
	return &Publisher{
		rdb: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
	}
}

func (p *Publisher) Close() error {
	if p == nil {
		return nil
	}
	return p.rdb.Close()
}

func (p *Publisher) PublishAgent(id uint, key string)     { p.publish(prefixAgent, id, key) }
func (p *Publisher) PublishCollector(id uint, key string) { p.publish(prefixCollector, id, key) }
func (p *Publisher) DeleteAgent(id uint)                  { p.delete(prefixAgent, id) }
func (p *Publisher) DeleteCollector(id uint)              { p.delete(prefixCollector, id) }

func (p *Publisher) publish(prefix string, id uint, key string) {
	if p == nil || key == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()

	if err := p.rdb.Set(ctx, prefix+fmt.Sprint(id), key, keyTTL).Err(); err != nil {
		// Not fatal: the input resolves a missing key by asking this service.
		_ = catcher.Error("cannot publish a connector key", err, map[string]any{
			"process": processName,
			"id":      id,
		})
	}
}

// delete is what makes revocation immediate. It is the one operation whose
// failure matters, because the key stays valid until it expires on its own.
func (p *Publisher) delete(prefix string, id uint) {
	if p == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()

	if err := p.rdb.Del(ctx, prefix+fmt.Sprint(id)).Err(); err != nil {
		_ = catcher.Error("cannot revoke a connector key, it stays valid until it expires", err, map[string]any{
			"process": processName,
			"id":      id,
			"expires": keyTTL.String(),
		})
	}
}

// Snapshot is what a republish pass hands over: every key that should exist.
type Snapshot struct {
	Agents     map[uint]string
	Collectors map[uint]string
}

// Run republishes on a timer until ctx ends.
func (p *Publisher) Run(ctx context.Context, snapshot func() Snapshot) {
	if p == nil {
		return
	}

	p.republish(ctx, snapshot())

	ticker := time.NewTicker(republishEvery)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.republish(ctx, snapshot())
		}
	}
}

func (p *Publisher) republish(ctx context.Context, s Snapshot) {
	cCtx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	pipe := p.rdb.Pipeline()
	for id, key := range s.Agents {
		pipe.Set(cCtx, prefixAgent+fmt.Sprint(id), key, keyTTL)
	}
	for id, key := range s.Collectors {
		pipe.Set(cCtx, prefixCollector+fmt.Sprint(id), key, keyTTL)
	}

	if _, err := pipe.Exec(cCtx); err != nil {
		_ = catcher.Error("cannot republish connector keys", err, map[string]any{
			"process":    processName,
			"agents":     len(s.Agents),
			"collectors": len(s.Collectors),
		})
	}
}
