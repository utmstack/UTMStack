package usecase

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/utmstack/utmstack/backend/modules/loganalyzer/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

func tenantCtx(id string) context.Context {
	return authz.WithTenantID(context.Background(), id)
}

func loaderReturning(calls *atomic.Int32, name string) func(context.Context) ([]dto.Field, error) {
	return func(context.Context) ([]dto.Field, error) {
		calls.Add(1)
		return []dto.Field{{Name: name}}, nil
	}
}

func TestFieldsCacheServesSecondCallWithoutLoading(t *testing.T) {
	var calls atomic.Int32
	c := newFieldsCache()
	load := loaderReturning(&calls, "a")

	for range 5 {
		if _, err := c.get(tenantCtx("t1"), "logs", load); err != nil {
			t.Fatalf("get: %v", err)
		}
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("expected 1 load, got %d", got)
	}
}

func TestFieldsCacheKeepsTenantsApart(t *testing.T) {
	var calls atomic.Int32
	c := newFieldsCache()

	a, err := c.get(tenantCtx("t1"), "logs", loaderReturning(&calls, "for-t1"))
	if err != nil {
		t.Fatalf("get t1: %v", err)
	}
	b, err := c.get(tenantCtx("t2"), "logs", loaderReturning(&calls, "for-t2"))
	if err != nil {
		t.Fatalf("get t2: %v", err)
	}
	if a[0].Name == b[0].Name {
		t.Fatalf("both tenants got %q", a[0].Name)
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("expected a load per tenant, got %d", got)
	}
}

// A stale entry must answer immediately rather than making the caller wait for
// the read it is trying to avoid.
func TestFieldsCacheServesStaleWhileRefreshing(t *testing.T) {
	var calls atomic.Int32
	c := newFieldsCache()
	ctx := tenantCtx("t1")

	if _, err := c.get(ctx, "logs", loaderReturning(&calls, "old")); err != nil {
		t.Fatalf("seed: %v", err)
	}
	c.mu.Lock()
	c.items["t1\x00logs"].loaded = time.Now().Add(-2 * fieldsTTL)
	c.mu.Unlock()

	released := make(chan struct{})
	slow := func(context.Context) ([]dto.Field, error) {
		<-released
		calls.Add(1)
		return []dto.Field{{Name: "new"}}, nil
	}

	done := make(chan []dto.Field, 1)
	go func() {
		got, err := c.get(ctx, "logs", slow)
		if err != nil {
			t.Errorf("get: %v", err)
		}
		done <- got
	}()

	select {
	case got := <-done:
		if got[0].Name != "old" {
			t.Fatalf("expected the stale value, got %q", got[0].Name)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("a stale entry made the caller wait for the refresh")
	}

	close(released)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		fresh, err := c.get(ctx, "logs", slow)
		if err != nil {
			t.Fatalf("get after refresh: %v", err)
		}
		if fresh[0].Name == "new" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("the background refresh never replaced the stale value")
}

// Several explorers opening at once must produce one read, not one each.
func TestFieldsCacheCollapsesConcurrentMisses(t *testing.T) {
	var calls atomic.Int32
	c := newFieldsCache()
	slow := func(context.Context) ([]dto.Field, error) {
		time.Sleep(50 * time.Millisecond)
		calls.Add(1)
		return []dto.Field{{Name: "a"}}, nil
	}

	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := c.get(tenantCtx("t1"), "logs", slow); err != nil {
				t.Errorf("get: %v", err)
			}
		}()
	}
	wg.Wait()

	if got := calls.Load(); got != 1 {
		t.Fatalf("expected 1 load for 8 concurrent callers, got %d", got)
	}
}

// A caller must not be able to corrupt what the next one reads.
func TestFieldsCacheReturnsACopy(t *testing.T) {
	var calls atomic.Int32
	c := newFieldsCache()
	load := loaderReturning(&calls, "original")

	first, err := c.get(tenantCtx("t1"), "logs", load)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	first[0].Name = "mutated"

	second, err := c.get(tenantCtx("t1"), "logs", load)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if second[0].Name != "original" {
		t.Fatalf("cache was mutated through a returned slice: %q", second[0].Name)
	}
}
