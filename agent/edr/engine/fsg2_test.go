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

// The HDB hash belongs only to the rebuilt child. A detection on the packed
// parent therefore proves that recursive scanning reached the FSG2 output.
func TestFSGRebuiltChildDetected(t *testing.T) {
	root := filepath.Join("..", "..", "..", "docs", "edr", "native-engine", "audits", "2026-10-26-fsg2", "oracle")
	type fixture struct{ packed, child, label string }
	cases := []fixture{
		{"fsg2-harmless.exe", "reference-rebuilt.exe", "FSG2.Rebuilt.Only"},
		{"fsg133-harmless.exe", "reference-133-rebuilt.exe", "FSG133.Rebuilt.Only"},
		{"fsg133-two-sections.exe", "reference-133-two-sections.exe", "FSG133.TwoSections.Only"},
	}
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	var rules string
	for _, tc := range cases {
		child, err := os.ReadFile(filepath.Join(root, tc.child))
		if err != nil {
			t.Fatal(err)
		}
		sum := md5.Sum(child)
		rules += fmt.Sprintf("%s:%d:%s\n", hex.EncodeToString(sum[:]), len(child), tc.label)
	}
	if err := os.WriteFile(filepath.Join(cfg.SignatureDBDir, "child.hdb"), []byte(rules), 0600); err != nil {
		t.Fatal(err)
	}
	e := New(cfg)
	if err := e.EnsureRunning(); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		packed, err := os.ReadFile(filepath.Join(root, tc.packed))
		if err != nil {
			t.Fatal(err)
		}
		clean, name, err := e.Scan(context.Background(), packed)
		if err != nil || clean || name != tc.label {
			t.Fatalf("packed %s: clean=%v, name=%q, err=%v", tc.packed, clean, name, err)
		}
	}
}
