package threadwinds

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/threatwinds/go-sdk/entities"

	"github.com/utmstack/utmstack/backend/modules/threatintel/contribution/models"
	"github.com/utmstack/utmstack/backend/modules/threatintel/contribution/service"
)

const sentKeyPrefix = "threatintel:contribution:sent:"

// KV is the part of the cache the registry needs. A cache that is missing or down answers "not
// there" and stores nothing, which makes every entity look new: the contribution then behaves as it
// does without a registry, sending everything every cycle.
type KV interface {
	Get(ctx context.Context, key string) (string, bool)
	Set(ctx context.Context, key, value string, ttl time.Duration)
}

// SentRegistry remembers which entities ThreatWinds already accepted, so a cycle does not send them
// again while they have not changed. An entity that changed hashes to something else and is sent.
type SentRegistry struct {
	kv  KV
	ttl time.Duration
}

// NewSentRegistry remembers each accepted entity for ttl.
func NewSentRegistry(kv KV, ttl time.Duration) *SentRegistry {
	return &SentRegistry{kv: kv, ttl: ttl}
}

// entityKey identifies an entity by what it says, not by the order its associations happened to be
// built in.
func entityKey(e *entities.Entity) (string, bool) {
	text, err := models.CanonicalJSON(e)
	if err != nil {
		return "", false
	}

	sum := sha256.Sum256([]byte(text))

	return sentKeyPrefix + hex.EncodeToString(sum[:]), true
}

// Wrap returns an ingestor that leaves out what was already sent and remembers what it sends.
func (r *SentRegistry) Wrap(inner service.Ingestor) service.Ingestor {
	return &skipSent{registry: r, inner: inner}
}

type skipSent struct {
	registry *SentRegistry
	inner    service.Ingestor
}

func (s *skipSent) IngestBatch(ctx context.Context, batch []*entities.Entity) models.IngestOutcome {
	fresh := make([]*entities.Entity, 0, len(batch))

	for _, e := range batch {
		if key, ok := entityKey(e); ok {
			if _, seen := s.registry.kv.Get(ctx, key); seen {
				continue
			}
		}

		fresh = append(fresh, e)
	}

	skipped := len(batch) - len(fresh)

	if len(fresh) == 0 {
		return models.IngestOutcome{Skipped: skipped}
	}

	outcome := s.inner.IngestBatch(ctx, fresh)
	outcome.Skipped = skipped

	for _, e := range outcome.Sent {
		if key, ok := entityKey(e); ok {
			s.registry.kv.Set(ctx, key, "1", s.registry.ttl)
		}
	}

	return outcome
}
