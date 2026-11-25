package threadwinds

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/threatwinds/go-sdk/entities"

	"github.com/utmstack/utmstack/backend/modules/threatintel/contribution/models"
)

type fakeKV struct {
	mu   sync.Mutex
	data map[string]string
	ttl  time.Duration
}

func newKV() *fakeKV { return &fakeKV{data: map[string]string{}} }

func (f *fakeKV) Get(_ context.Context, k string) (string, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v, ok := f.data[k]

	return v, ok
}

func (f *fakeKV) Set(_ context.Context, k, v string, ttl time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.data[k], f.ttl = v, ttl
}

// nilKV behaves like the SDK cache without an address: always a miss, stores nothing.
type nilKV struct{}

func (nilKV) Get(context.Context, string) (string, bool)         { return "", false }
func (nilKV) Set(context.Context, string, string, time.Duration) {}

// recorder is an inner ingestor that accepts everything except what it is told to fail.
type recorder struct {
	got  [][]*entities.Entity
	fail map[string]bool
}

func (r *recorder) IngestBatch(_ context.Context, batch []*entities.Entity) models.IngestOutcome {
	r.got = append(r.got, batch)

	out := models.IngestOutcome{}

	for _, e := range batch {
		if r.fail[e.Type] {
			out.Failures = append(out.Failures, models.EntityIngestResult{Entity: e})

			continue
		}

		out.OKCount++
		out.Sent = append(out.Sent, e)
	}

	return out
}

func ent(typ, value string) *entities.Entity {
	return &entities.Entity{Type: typ, Tags: []string{value}}
}

func TestWhatWasSentIsNotSentAgain(t *testing.T) {
	kv, inner := newKV(), &recorder{}
	ing := NewSentRegistry(kv, time.Hour).Wrap(inner)

	first := ing.IngestBatch(context.Background(), []*entities.Entity{ent("ip", "1"), ent("domain", "a")})
	if first.OKCount != 2 || first.Skipped != 0 {
		t.Fatalf("first: %+v", first)
	}

	second := ing.IngestBatch(context.Background(), []*entities.Entity{ent("ip", "1"), ent("domain", "a")})
	if second.Skipped != 2 || second.OKCount != 0 || len(inner.got) != 1 {
		t.Errorf("second: %+v, inner calls %d", second, len(inner.got))
	}

	if kv.ttl != time.Hour {
		t.Errorf("ttl %v", kv.ttl)
	}
}

func TestAnEntityThatChangedIsSentAgain(t *testing.T) {
	inner := &recorder{}
	ing := NewSentRegistry(newKV(), time.Hour).Wrap(inner)

	ing.IngestBatch(context.Background(), []*entities.Entity{ent("ip", "1")})
	out := ing.IngestBatch(context.Background(), []*entities.Entity{ent("ip", "1"), ent("ip", "2")})

	if out.Skipped != 1 || out.OKCount != 1 {
		t.Errorf("%+v", out)
	}

	if last := inner.got[len(inner.got)-1]; len(last) != 1 || last[0].Tags[0] != "2" {
		t.Errorf("the inner ingestor should only get the new entity: %+v", last)
	}
}

func TestFailedEntitiesAreNotRemembered(t *testing.T) {
	inner := &recorder{fail: map[string]bool{"ip": true}}
	ing := NewSentRegistry(newKV(), time.Hour).Wrap(inner)

	ing.IngestBatch(context.Background(), []*entities.Entity{ent("ip", "1")})

	inner.fail = nil

	out := ing.IngestBatch(context.Background(), []*entities.Entity{ent("ip", "1")})
	if out.OKCount != 1 || out.Skipped != 0 {
		t.Errorf("a failed entity must be retried: %+v", out)
	}
}

func TestWithoutACacheEverythingIsSentEveryTime(t *testing.T) {
	inner := &recorder{}
	ing := NewSentRegistry(nilKV{}, time.Hour).Wrap(inner)

	for i := 0; i < 2; i++ {
		if out := ing.IngestBatch(context.Background(), []*entities.Entity{ent("ip", "1")}); out.OKCount != 1 || out.Skipped != 0 {
			t.Errorf("round %d: %+v", i, out)
		}
	}
}

func TestAnAllSkippedBatchDoesNotCallTheInnerIngestor(t *testing.T) {
	kv, inner := newKV(), &recorder{}
	ing := NewSentRegistry(kv, time.Hour).Wrap(inner)

	ing.IngestBatch(context.Background(), []*entities.Entity{ent("ip", "1")})
	ing.IngestBatch(context.Background(), []*entities.Entity{ent("ip", "1")})

	if len(inner.got) != 1 {
		t.Errorf("inner called %d times", len(inner.got))
	}
}
