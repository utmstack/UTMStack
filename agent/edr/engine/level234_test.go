package engine

import (
	"context"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func TestPinnedLevel234SelectionMatchesNativeOracle(t *testing.T) {
	if signatures.FunctionalityLevel != 234 {
		t.Fatalf("functionality level = %d, want pinned ClamAV 1.5.4 level 234", signatures.FunctionalityLevel)
	}
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	_, err := e.ActivateGeneration(map[string][]byte{
		"level.ndb": []byte("Native.Level234:0:*:41424344:234\nNative.Level235:0:*:45464748:235\n"),
	}, 1, manifestHash("level234-native-oracle"))
	if err != nil {
		t.Fatal(err)
	}
	if got := e.current.Load().db.Report().Skipped[signatures.SkipLevel]; got != 1 {
		t.Fatalf("level-235 rule skip count = %d, want 1", got)
	}
	for _, tc := range []struct {
		data []byte
		want string
	}{
		{[]byte("ABCDabcd"), "Native.Level234"},
		{[]byte("EFGHefgh"), ""},
	} {
		clean, name, err := e.Scan(context.Background(), tc.data)
		if err != nil || clean != (tc.want == "") || name != tc.want {
			t.Fatalf("%q: clean=%v name=%q err=%v, want %q", tc.data, clean, name, err, tc.want)
		}
	}
}
