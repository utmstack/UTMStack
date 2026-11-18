//go:build linux || darwin

package quarantine

import (
	"os"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"

	"github.com/google/uuid"
)

func TestInspectRecoveryRefusesUnsafeDirectoriesWithoutHardening(t *testing.T) {
	for _, unsafe := range []string{"root", "journal"} {
		t.Run(unsafe, func(t *testing.T) {
			s, _, _, _ := admissionFixture(t)
			path := s.dir
			if unsafe == "journal" {
				path = filepath.Join(s.dir, ".restores")
				if err := os.Mkdir(path, 0700); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Chmod(path, 0755); err != nil {
				t.Fatal(err)
			}
			before := inspectionTree(t, s.dir)
			if _, err := InspectRecovery(s.dir, uuid.NewString()); err == nil {
				t.Fatal("accepted non-private recovery directory")
			}
			if after := inspectionTree(t, s.dir); !reflect.DeepEqual(before, after) {
				t.Fatal("inspection hardened or changed an unsafe directory")
			}
		})
	}
}

func TestInspectRecoveryRefusesFIFOWithoutBlocking(t *testing.T) {
	s, _, _, _ := admissionFixture(t)
	id := uuid.NewString()
	op := filepath.Join(s.dir, ".restores", id)
	if err := os.MkdirAll(op, 0700); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(filepath.Join(op, "manifest.json"), 0600); err != nil {
		t.Fatal(err)
	}
	report, err := InspectRecovery(s.dir, id)
	if err == nil || report.Journals[1].Status != "unreadable" {
		t.Fatalf("non-regular journal: %+v %v", report, err)
	}
}
