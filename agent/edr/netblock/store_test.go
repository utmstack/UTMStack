// edr/netblock/store_test.go
package netblock

import (
	"net/netip"
	"sync"
	"testing"
)

func TestStoreSwapAtomic(t *testing.T) {
	s := NewStore()
	i1, _ := ParseIndicator("1.2.3.4", "ip", 1)
	s.Swap([]Indicator{i1})
	if _, ok := s.Match(netip.MustParseAddr("1.2.3.4")); !ok {
		t.Fatal("expected match after first swap")
	}
	// Concurrent readers during a swap must never see a torn state.
	var wg sync.WaitGroup
	stop := make(chan struct{})
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					s.Match(netip.MustParseAddr("5.6.7.8"))
				}
			}
		}()
	}
	i2, _ := ParseIndicator("5.6.7.8", "ip", 1)
	for i := 0; i < 100; i++ {
		s.Swap([]Indicator{i1, i2})
		s.Swap([]Indicator{i1})
	}
	close(stop)
	wg.Wait()
	if s.Count() != 1 {
		t.Fatalf("final count = %d, want 1", s.Count())
	}
}
