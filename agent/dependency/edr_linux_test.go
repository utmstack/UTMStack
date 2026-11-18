//go:build linux

package dependency

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLinuxNativeDependencyInventoryAndStagingPermissions(t *testing.T) {
	seen := map[string]int{}
	for _, dep := range GetDependencies() {
		seen[dep.Name]++
		if seen[dep.Name] > 1 {
			t.Fatalf("duplicate dependency %s", dep.Name)
		}
	}
	if seen["edr"] != 1 || seen["auditd"] != 1 {
		t.Fatalf("missing native Linux integration: %v", seen)
	}
	root := t.TempDir()
	staged := filepath.Join(root, "incoming")
	if err := os.WriteFile(staged, []byte("inert placeholder, never executed"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := prepareEDRExecutable(staged); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(staged)
	if err != nil || st.Mode().Perm() != 0755 {
		t.Fatalf("executable permissions %v %v", st, err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(staged, link); err != nil {
		t.Fatal(err)
	}
	if err := prepareEDRExecutable(link); err == nil {
		t.Fatal("staged symlink accepted")
	}
}
