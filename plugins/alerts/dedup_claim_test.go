package main

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/store/cache"
)

func newTestKV(t *testing.T) *cache.Client {
	t.Helper()
	mr := miniredis.RunT(t)
	return cache.New(cache.Config{Addr: mr.Addr()})
}

// ALR-F19: a dedup slot someone else already claimed must make isDuplicate
// return true immediately, with no ClickHouse round trip — alertStore stays
// nil here, so a bug that let the call fall through would panic the test,
// not just fail an assertion.
func TestIsDuplicateReturnsTrueOnLostClaimWithoutTouchingTheStore(t *testing.T) {
	kv = newTestKV(t)
	t.Cleanup(func() { kv = nil })

	alert := &plugins.Alert{
		Id:            "second-event",
		Name:          "brute-force-ssh",
		TenantId:      "tenant-1",
		DeduplicateBy: []string{"name"},
	}
	alertJSON := `{"name":"brute-force-ssh"}`

	token := deduplicationToken(alert, alertJSON)
	if token == "" {
		t.Fatal("expected a non-empty dedup token for this alert")
	}

	// Simulate the first of two events arriving together: it wins the claim
	// first.
	claimed, _, err := kv.Claim(context.Background(), dedupCachePrefix+token, "first-event", fastPathTTL)
	if err != nil || !claimed {
		t.Fatalf("setup: first claim = (%v, %v), want (true, nil)", claimed, err)
	}

	if !isDuplicate(alert, alertJSON) {
		t.Fatal("isDuplicate() = false for an alert whose dedup slot is already claimed by someone else, want true")
	}
}

// Two events with the same DeduplicateBy key arriving at the same instant
// must not both be told "not a duplicate" — exactly one proceeds.
func TestIsDuplicateClaimPicksExactlyOneWinner(t *testing.T) {
	kv = newTestKV(t)
	t.Cleanup(func() { kv = nil })

	token := deduplicationToken(&plugins.Alert{DeduplicateBy: []string{"name"}}, `{"name":"brute-force-ssh"}`)

	const n = 20
	var wins int32
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			claimed, _, err := kv.Claim(context.Background(), dedupCachePrefix+token, "claimant", time.Minute)
			if err == nil && claimed {
				mu.Lock()
				wins++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()

	if wins != 1 {
		t.Fatalf("exactly-one-winner race: %d goroutines won the claim, want 1", wins)
	}
}
