package threadwinds

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/threatwinds/go-sdk/entities"

	"github.com/utmstack/utmstack/backend/pkg/instanceconfig"
)

func fast(t *testing.T) {
	t.Helper()

	oldSleep, oldBackoff := entitySleep, initialBackof
	entitySleep, initialBackof = time.Millisecond, time.Millisecond

	t.Cleanup(func() { entitySleep, initialBackof = oldSleep, oldBackoff })
}

func batchOf(n int) []*entities.Entity {
	out := make([]*entities.Entity, n)
	for i := range out {
		out[i] = &entities.Entity{Type: "domain", Tags: []string{string(rune('a' + i))}}
	}

	return out
}

func clientFor(srv *httptest.Server) *Client {
	return New(func() *instanceconfig.InstanceConfig {
		return &instanceconfig.InstanceConfig{Server: srv.URL, InstanceID: "inst-1", InstanceKey: "secret"}
	})
}

func TestAcceptedEntitiesAreSentWithTheInstanceCredentials(t *testing.T) {
	fast(t)

	var path, id, key string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, id, key = r.URL.Path, r.Header.Get("id"), r.Header.Get("key")
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	out := clientFor(srv).IngestBatch(context.Background(), batchOf(2))

	if out.OKCount != 2 || len(out.Sent) != 2 || len(out.Failures) != 0 {
		t.Fatalf("outcome: %+v", out)
	}

	if path != "/proxy/api/ingest/v1/entity" || id != "inst-1" || key != "secret" {
		t.Errorf("path=%q id=%q key=%q", path, id, key)
	}
}

func TestAnUnprovisionedInstanceStopsTheBatch(t *testing.T) {
	fast(t)

	for _, status := range []int{http.StatusUnauthorized, http.StatusUnprocessableEntity} {
		var calls int32

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&calls, 1)
			w.WriteHeader(status)
		}))

		out := clientFor(srv).IngestBatch(context.Background(), batchOf(3))
		srv.Close()

		if !out.ProvisioningBlocked || calls != 1 || out.OKCount != 0 || len(out.Failures) != 1 {
			t.Errorf("status %d: calls=%d outcome=%+v", status, calls, out)
		}
	}
}

func TestANotFoundIsNotRetriedAndDoesNotStopTheBatch(t *testing.T) {
	fast(t)

	var calls int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	out := clientFor(srv).IngestBatch(context.Background(), batchOf(2))

	if calls != 2 || out.ProvisioningBlocked || len(out.Failures) != 2 || len(out.Sent) != 0 {
		t.Errorf("calls=%d outcome=%+v", calls, out)
	}
}

func TestServerErrorsAreRetriedUntilTheyClear(t *testing.T) {
	fast(t)

	var calls int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) < 3 {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	out := clientFor(srv).IngestBatch(context.Background(), batchOf(1))

	if calls != 3 || out.OKCount != 1 {
		t.Errorf("calls=%d outcome=%+v", calls, out)
	}
}

func TestServerErrorsThatNeverClearFailTheEntityAfterThreeAttempts(t *testing.T) {
	fast(t)

	var calls int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	out := clientFor(srv).IngestBatch(context.Background(), batchOf(1))

	if calls != 3 || len(out.Failures) != 1 || len(out.Sent) != 0 || out.ProvisioningBlocked {
		t.Errorf("calls=%d outcome=%+v", calls, out)
	}
}

func TestAnInstanceWithoutCredentialsSendsNothing(t *testing.T) {
	fast(t)

	for name, cfg := range map[string]*instanceconfig.InstanceConfig{
		"none":      nil,
		"no server": {InstanceID: "i", InstanceKey: "k"},
		"no id":     {Server: "http://x", InstanceKey: "k"},
		"no key":    {Server: "http://x", InstanceID: "i"},
	} {
		c := New(func() *instanceconfig.InstanceConfig { return cfg })

		out := c.IngestBatch(context.Background(), batchOf(1))
		if !out.ProvisioningBlocked || out.OKCount != 0 {
			t.Errorf("%s: %+v", name, out)
		}
	}
}

func TestACancelledContextStopsTheBatch(t *testing.T) {
	fast(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusAccepted) }))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if out := clientFor(srv).IngestBatch(ctx, batchOf(3)); out.OKCount != 0 {
		t.Errorf("%+v", out)
	}
}
