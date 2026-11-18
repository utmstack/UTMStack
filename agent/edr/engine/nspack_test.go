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

func TestNsPackRebuiltChildDetected(t *testing.T) {
	root := filepath.Join("..", "..", "..", "docs", "edr", "native-engine", "audits", "2026-10-26-nspack", "oracle")
	cases := []struct{ packed, child, label string }{
		{"nspack-zero.exe", "reference-rebuilt.exe", "NsPack.Rebuilt.Only"},
		{"nspack-mutation-128.exe", "reference-mutation-128.exe", "NsPack.Mutated.Only"},
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
