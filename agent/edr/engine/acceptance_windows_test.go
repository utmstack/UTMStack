//go:build windows && edr_acceptance

package engine

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"os"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

// Memory-only public EICAR data cannot be removed by a competing filesystem
// scanner before the owned engine sees it. No scanned program is executed.
func TestWindowsAcceptanceOfficialFeedBuffers(t *testing.T) {
	if os.Getenv("EDR_WINDOWS_ACCEPTANCE") != "isolated-vm" {
		t.Skip("requires explicit disposable Windows acceptance")
	}
	dir := os.Getenv("EDR_ACCEPTANCE_OFFICIAL_FEED")
	if dir == "" {
		t.Fatal("current acceptance requires the pinned official feed")
	}
	e := New(config.EDRConfig{SignatureDBDir: dir})
	if err := e.EnsureRunning(); err != nil {
		t.Fatal(err)
	}
	defer e.Stop()
	if e.SignatureCount() < 3_000_000 {
		t.Fatalf("full feed not loaded: %d", e.SignatureCount())
	}
	positive, err := base64.StdEncoding.DecodeString("WDVPIVAlQEFQWzRcUFpYNTQoUF4pN0NDKTd9JEVJQ0FSLVNUQU5EQVJELUFOVElWSVJVUy1URVNULUZJTEUhJEgrSCo=")
	if err != nil {
		t.Fatal(err)
	}
	wrap := func(data []byte) []byte {
		var output bytes.Buffer
		w := zip.NewWriter(&output)
		entry, err := w.Create("public-test-data.txt")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write(data); err != nil {
			t.Fatal(err)
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
		return output.Bytes()
	}
	for _, tc := range []struct {
		name      string
		data      []byte
		wantClean bool
	}{
		{"public-eicar", positive, false},
		{"public-eicar-zip", wrap(positive), false},
		{"public-eicar-nested-zip", wrap(wrap(positive)), false},
		{"benign", []byte("Ordinary harmless OpenEDR Windows acceptance text."), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clean, name, scanErr := e.Scan(context.Background(), tc.data)
			if scanErr != nil || clean != tc.wantClean || (!tc.wantClean && name == "") {
				t.Fatalf("clean=%v signature=%q error=%v wantClean=%v", clean, name, scanErr, tc.wantClean)
			}
			t.Logf("implementation=%s generation=%s signatures=%d clean=%v signature=%q", ImplementationVersion, e.Version(), e.SignatureCount(), clean, name)
		})
	}
}
