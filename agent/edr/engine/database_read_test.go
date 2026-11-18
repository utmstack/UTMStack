package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBoundedDatabaseRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "database.ndb")
	if err := os.WriteFile(path, []byte("abc"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, limit := range []int64{-1, 0, 2} {
		if _, err := readBoundedFile(path, limit); err == nil {
			t.Fatalf("accepted limit %d", limit)
		}
	}
	if got, err := readBoundedFile(path, 3); err != nil || string(got) != "abc" {
		t.Fatalf("read: %q, %v", got, err)
	}
	if _, err := readBoundedFile(dir, 1<<20); err == nil {
		t.Fatal("accepted nonregular database input")
	}
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if got, err := readBoundedFile(path, 0); err != nil || len(got) != 0 {
		t.Fatalf("empty read: %q, %v", got, err)
	}
}
