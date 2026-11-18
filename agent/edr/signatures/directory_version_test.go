package signatures

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirectoryContentVersionMatchesCompilerIdentity(t *testing.T) {
	dir := t.TempDir()
	files := map[string][]byte{
		"a.ndb": []byte("First:0:*:41424344\n"),
		"b.ldb": []byte("Second;Target:0;0;44434241\n"),
	}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "state.json"), []byte("ignored"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "generations"), 0o700); err != nil {
		t.Fatal(err)
	}
	want, err := ContentVersion(files, Limits{})
	if err != nil {
		t.Fatal(err)
	}
	got, err := DirectoryContentVersion(dir, Limits{})
	if err != nil || got != want {
		t.Fatalf("directory version %q, want %q: %v", got, want, err)
	}
	changed := []byte("First:0:*:44434241\n") // Equal length, different bytes.
	if err := os.WriteFile(filepath.Join(dir, "a.ndb"), changed, 0o600); err != nil {
		t.Fatal(err)
	}
	if got, err := DirectoryContentVersion(dir, Limits{}); err != nil || got == want {
		t.Fatalf("equal-size content change retained identity: %q, %v", got, err)
	}
	if _, err := DirectoryContentVersion(dir, Limits{MaxFileBytes: 4}); err == nil {
		t.Fatal("oversized file accepted")
	}
	if _, err := DirectoryContentVersion(dir, Limits{MaxFiles: 1}); err == nil {
		t.Fatal("excess file count accepted")
	}
	if err := os.Symlink("a.ndb", filepath.Join(dir, "alias.ndb")); err == nil {
		if _, err := DirectoryContentVersion(dir, Limits{}); err == nil {
			t.Fatal("symbolic link accepted")
		}
	}
}
