package engine

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

// Each fixture has a recorded ClamAV 1.5.4 verdict in the adjacent audit.
// The ZIP-overlay case also checks that its child detection wins over the
// parent's broken-PE heuristic in both modes.
func TestBrokenPEHeaderReferenceModes(t *testing.T) {
	root := filepath.Join("..", "..", "..", "docs", "edr", "native-engine", "audits", "2026-10-27-v127-broken-pe", "oracle")
	marker := []byte("UTMStack inert overlay marker")
	sum := md5.Sum(marker)
	rules := fmt.Sprintf("%s:%d:Overlay.Marker\n", hex.EncodeToString(sum[:]), len(marker))
	// A second unrelated rule keeps a nonempty database for fixtures without
	// any embedded marker.
	rules += "900150983cd24fb0d6963f7d28e17f72:3:Dummy\n"
	db := t.TempDir()
	if err := os.WriteFile(filepath.Join(db, "probe.hdb"), []byte(rules), 0600); err != nil {
		t.Fatal(err)
	}
	strict := config.Default()
	strict.SignatureDBDir = db
	strictEngine := New(strict)
	if err := strictEngine.EnsureRunning(); err != nil {
		t.Fatal(err)
	}
	permissive := strict
	off := false
	permissive.ScanHeuristics.BrokenExecutable = &off
	permissiveEngine := New(permissive)
	if err := permissiveEngine.EnsureRunning(); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"first-gap", "misaligned-rva", "one-section-past-end",
		"section-gap", "section-overlap", "short-dos", "short-optional", "sparse-pe",
		"unmapped-entry", "zero-sections", "valid", "not-pe", "sparse-pe-zip-overlay"} {
		b, err := os.ReadFile(filepath.Join(root, name+".exe"))
		if err != nil {
			t.Fatal(err)
		}
		wantStrict := "Heuristics.Broken.Executable"
		if name == "valid" || name == "not-pe" {
			wantStrict = ""
		}
		wantPermissive := ""
		if name == "sparse-pe-zip-overlay" {
			wantStrict, wantPermissive = "Overlay.Marker", "Overlay.Marker"
		}
		for _, tc := range []struct {
			mode, want string
			engine     *Engine
		}{{"strict", wantStrict, strictEngine}, {"permissive", wantPermissive, permissiveEngine}} {
			t.Run(name+"/"+tc.mode, func(t *testing.T) {
				clean, got, err := tc.engine.Scan(context.Background(), b)
				if err != nil || got != tc.want || clean != (tc.want == "") {
					t.Fatalf("clean=%v name=%q err=%v, want %q", clean, got, err, tc.want)
				}
			})
		}
	}
}
