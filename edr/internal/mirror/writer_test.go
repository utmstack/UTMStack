// utmstack-v12/edr/internal/mirror/writer_test.go
package mirror

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteArtifactWritesDataAndChecksumSibling(t *testing.T) {
	root := t.TempDir()
	data := []byte("1.2.3.4\n5.6.7.8\n")
	rel := "feeds/v1/download/list/level1/accumulative/ip"
	if err := WriteArtifact(root, rel, data); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil || string(got) != string(data) {
		t.Fatalf("artifact = %q, err %v", got, err)
	}
	sum, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)+".sha256"))
	if err != nil {
		t.Fatal(err)
	}
	want := sha256.Sum256(data)
	if string(sum) != hex.EncodeToString(want[:]) {
		t.Fatalf("checksum = %q, want %q", sum, hex.EncodeToString(want[:]))
	}
}

func TestWriteArtifactOverwritesAndLeavesNoTempFiles(t *testing.T) {
	root := t.TempDir()
	if err := WriteArtifact(root, "a/b", []byte("old")); err != nil {
		t.Fatal(err)
	}
	if err := WriteArtifact(root, "a/b", []byte("new")); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(root, "a", "b"))
	if string(got) != "new" {
		t.Fatalf("artifact = %q, want new", got)
	}
	entries, _ := os.ReadDir(filepath.Join(root, "a"))
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Fatalf("temp file left behind: %s", e.Name())
		}
	}
}
