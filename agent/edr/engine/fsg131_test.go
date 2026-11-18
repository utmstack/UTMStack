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

// The HDB digest is absent from the packed parent and belongs to the
// reference's rebuilt child, proving recursive scanning reached that child.
func TestFSG131RebuiltChildDetected(t *testing.T) {
	root := filepath.Join("..", "..", "..", "docs", "edr", "native-engine", "audits", "2026-10-27-fsg131", "oracle")
	cfg := config.Default()
	off := false
	cfg.ScanHeuristics.BrokenExecutable = &off
	cfg.SignatureDBDir = t.TempDir()
	var rules string
	for _, tc := range []struct{ name, label string }{
		{"fsg131-harmless", "FSG131.Rebuilt.Only"},
		{"fsg131-two-sections", "FSG131.TwoSections.Only"},
	} {
		child, err := os.ReadFile(filepath.Join(root, tc.name+"-reference-rebuilt.exe"))
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
	strict := config.Default()
	strict.SignatureDBDir = cfg.SignatureDBDir
	strictEngine := New(strict)
	if err := strictEngine.EnsureRunning(); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, label string }{
		{"fsg131-harmless", "FSG131.Rebuilt.Only"},
		{"fsg131-two-sections", "FSG131.TwoSections.Only"},
		{"fsg131-call-stub", "FSG131.Rebuilt.Only"},
	} {
		packed, err := os.ReadFile(filepath.Join(root, tc.name+".exe"))
		if err != nil {
			t.Fatal(err)
		}
		clean, name, err := e.Scan(context.Background(), packed)
		if err != nil || clean || name != tc.label {
			t.Errorf("%s permissive scan: clean=%v, name=%q, err=%v", tc.name, clean, name, err)
		}
		clean, name, err = strictEngine.Scan(context.Background(), packed)
		if err != nil || clean || name != "Heuristics.Broken.Executable" {
			t.Errorf("%s strict scan: clean=%v, name=%q, err=%v", tc.name, clean, name, err)
		}
	}
}
