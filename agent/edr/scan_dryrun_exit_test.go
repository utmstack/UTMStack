package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestDryRunExitDistinguishesIncompleteFromClean(t *testing.T) {
	db := t.TempDir()
	if err := os.WriteFile(filepath.Join(db, "test.ndb"), []byte("Audit.Marker:0:*:61756469746d61726b6572\n"), 0600); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	for name, content := range map[string]string{
		"clean":      "ordinary benign content",
		"incomplete": "ISc(unsupported cabinet",
		"detected":   "auditmarker",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, tt := range []struct {
		name  string
		paths []string
		code  int
	}{
		{"clean", []string{"clean"}, 0},
		{"incomplete", []string{"incomplete"}, 2},
		{"missing", []string{"missing"}, 2},
		{"clean and incomplete", []string{"clean", "incomplete"}, 2},
		{"detection and incomplete", []string{"detected", "incomplete"}, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{"--dry-run", "--json", "--db", db}
			for _, p := range tt.paths {
				args = append(args, filepath.Join(dir, p))
			}
			var out bytes.Buffer
			if code := runDryRun(args, &out); code != tt.code {
				t.Fatalf("exit=%d, want %d: %s", code, tt.code, out.String())
			}
		})
	}
}
