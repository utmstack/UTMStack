//go:build !windows

package ransomware

import (
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

// TestManager_LoadRehydratesCaseSensitive pins the POSIX Load rule: a
// rehydrated membership set matches a feed path only when the case agrees
// with the stored path — a case-different path is a different file and must
// NOT count as the canary.
func TestManager_LoadRehydratesCaseSensitive(t *testing.T) {
	fs := &fakeCanaryStore{recs: []cache.CanaryRecord{{Path: `/home/u/docs/00__accounts.xlsx`}}}
	m := NewManager(fs)
	if err := m.Load(); err != nil {
		t.Fatal(err)
	}
	if !m.Contains(`/home/u/docs/00__accounts.xlsx`) {
		t.Error("same-case path must match the planted canary")
	}
	if m.Contains(`/home/u/docs/00__Accounts.xlsx`) {
		t.Error("case-different path must NOT match on a case-sensitive filesystem")
	}
}
