package coordination

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

const (
	// LeaseBucketName is exported so plugins can name the bucket in their logs
	// without keeping a copy of the string.
	// Keys are prefixed with the plugin name, e.g. "aws.<tenantId>/<group>".
	LeaseBucketName = "ingest_leases"

	// Bucket-level, not per key: per-key TTL requires NATS 2.11+ and the
	// deployed server is nats:2.10-alpine.
	bucketTTL = 45 * time.Second

	bucketHistory = 1

	// Separate bucket: one bucket cannot hold both TTLs.
	schedulerBucketName = "ingest_scheduler"

	schedulerBucketTTL = 20 * time.Second
)

type leaseValue struct {
	Holder     string    `json:"holder"`
	AcquiredAt time.Time `json:"acquiredAt"`
}

type NATSStore struct {
	kv    jetstream.KeyValue
	clock Clock
}

func NewNATSStore(ctx context.Context, js jetstream.JetStream, clock Clock) (*NATSStore, error) {
	return NewNATSStoreWithBucket(ctx, js, clock, LeaseBucketName, bucketTTL)
}

func NewNATSSchedulerStore(ctx context.Context, js jetstream.JetStream, clock Clock) (*NATSStore, error) {
	return NewNATSStoreWithBucket(ctx, js, clock, schedulerBucketName, schedulerBucketTTL)
}

// Expiry on NATS is a property of the bucket, not the clock, so a caller
// needing a different TTL needs a different bucket.
func NewNATSStoreWithBucket(ctx context.Context, js jetstream.JetStream, clock Clock, bucket string, ttl time.Duration) (*NATSStore, error) {
	kv, err := js.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:  bucket,
		TTL:     ttl,
		History: bucketHistory,
	})
	if err != nil {
		return nil, err
	}

	return &NATSStore{kv: kv, clock: clock}, nil
}

func (s *NATSStore) Acquire(ctx context.Context, key, holder string, ttl time.Duration) (Lease, error) {
	acquiredAt := s.clock.Now()
	data, err := json.Marshal(leaseValue{Holder: holder, AcquiredAt: acquiredAt})
	if err != nil {
		return Lease{}, err
	}

	revision, err := s.kv.Create(ctx, key, data)
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyExists) {
			return Lease{}, ErrHeld
		}
		return Lease{}, err
	}

	return Lease{
		Key:       key,
		Holder:    holder,
		Revision:  revision,
		ExpiresAt: acquiredAt.Add(ttl),
	}, nil
}

func (s *NATSStore) Renew(ctx context.Context, l Lease, ttl time.Duration) (Lease, error) {
	renewedAt := s.clock.Now()
	data, err := json.Marshal(leaseValue{Holder: l.Holder, AcquiredAt: renewedAt})
	if err != nil {
		return Lease{}, err
	}

	revision, err := s.kv.Update(ctx, l.Key, data, l.Revision)
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyRevisionMismatch) {
			return Lease{}, ErrFenced
		}
		return Lease{}, err
	}

	return Lease{
		Key:       l.Key,
		Holder:    l.Holder,
		Revision:  revision,
		ExpiresAt: renewedAt.Add(ttl),
	}, nil
}

// Revision-guarded on purpose: unguarded, a holder that froze past its TTL and
// then called Release would delete the lease a successor had already won.
func (s *NATSStore) Release(ctx context.Context, l Lease) error {
	err := s.kv.Delete(ctx, l.Key, jetstream.LastRevision(l.Revision))
	if errors.Is(err, jetstream.ErrKeyRevisionMismatch) {
		return ErrFenced
	}
	return err
}
