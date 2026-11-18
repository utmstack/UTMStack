package service

import (
	"os"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"
)

func TestOptionalProfileCanariesSkipOnlyAbsentFolders(t *testing.T) {
	profile := t.TempDir()
	existing := filepath.Join(profile, "Documents")
	if err := os.Mkdir(existing, 0700); err != nil {
		t.Fatal(err)
	}
	// Existing non-directory/symlink objects must reach the pinned directory
	// validator, which reports the actual problem instead of claiming coverage.
	file := filepath.Join(profile, "Desktop")
	if err := os.WriteFile(file, []byte("ordinary non-directory control"), 0600); err != nil {
		t.Fatal(err)
	}
	got := optionalProfileCanaryDirs(profile, os.Lstat)
	want := []string{file, existing}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("optional scope = %q, want %q", got, want)
	}
}

func TestOptionalProfileCanariesRetainAccessAndSymlinkFailures(t *testing.T) {
	profile := t.TempDir()
	target := t.TempDir()
	link := filepath.Join(profile, "Downloads")
	if err := os.Symlink(target, link); err != nil {
		t.Skip("symlink fixture unavailable", err)
	}
	if got := optionalProfileCanaryDirs(profile, os.Lstat); !reflect.DeepEqual(got, []string{link}) {
		t.Fatal(got)
	}
	denied := func(name string) (os.FileInfo, error) {
		return nil, &os.PathError{Op: "lstat", Path: name, Err: syscall.EACCES}
	}
	got := optionalProfileCanaryDirs(profile, denied)
	if len(got) != 4 {
		t.Fatalf("permission failures were hidden: %q", got)
	}
	for _, path := range got {
		if filepath.Dir(path) != profile {
			t.Fatal("out of scope", path)
		}
	}
	if got := optionalProfileCanaryDirs("", denied); len(got) != 0 {
		t.Fatal("empty profile expanded", got)
	}
}
