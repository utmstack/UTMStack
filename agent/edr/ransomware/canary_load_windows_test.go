//go:build windows

package ransomware

import (
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

// TestManager_LoadRehydratesCaseInsensitive pins the Windows Load rule:
// membership is case-insensitive, so a feed path differing only in case from
// the stored canary path is the same file.
func TestManager_LoadRehydratesCaseInsensitive(t *testing.T) {
	fs := &fakeCanaryStore{recs: []cache.CanaryRecord{{Path: `C:\Users\x\00__a.xlsx`}}}
	m := NewManager(fs)
	if err := m.Load(); err != nil {
		t.Fatal(err)
	}
	if !m.Contains(`c:\users\x\00__A.xlsx`) {
		t.Error("case-different path must match on Windows")
	}
}
