//go:build windows

package quarantine

import (
	"errors"
	"os"
	"testing"

	"golang.org/x/sys/windows"
)

func TestRestoreWindowsRejectsRetainedWriter(t *testing.T) {
	s, _, id, original, stored := restoreFixture(t)
	writer, err := os.OpenFile(stored, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Restore(id); err == nil {
		writer.Close()
		t.Fatal("retained writer allowed restoration")
	}
	if _, err := os.Stat(original); !os.IsNotExist(err) {
		writer.Close()
		t.Fatal("busy file published", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := s.Restore(id); err != nil {
		t.Fatalf("retry after writer close failed: %v", err)
	}
}

func TestRestoreWindowsMappedSourceCannotChangePublishedSnapshot(t *testing.T) {
	s, c, id, original, stored := restoreFixture(t)
	writer, err := os.OpenFile(stored, os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	mapping, err := windows.CreateFileMapping(windows.Handle(writer.Fd()), nil, windows.PAGE_READWRITE, 0, 0, nil)
	if err != nil {
		writer.Close()
		t.Fatal(err)
	}
	defer func() {
		if mapping != 0 {
			windows.CloseHandle(mapping)
		}
	}()
	view, err := windows.MapViewOfFile(mapping, windows.FILE_MAP_WRITE, 0, 0, 0)
	if err != nil {
		writer.Close()
		t.Fatal(err)
	}
	defer func() {
		if view != 0 {
			windows.UnmapViewOfFile(view)
		}
	}()
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	// Exercise the owned mapping writer without changing the starting bytes.
	writeOwnedRestoreMapping(t, view, []byte("payload"))
	restoreErr := s.Restore(id)
	if restoreErr == nil {
		t.Fatal("mapped source retirement failure was not reported")
	}
	if _, statErr := os.Stat(original); os.IsNotExist(statErr) {
		// Refusal before publication while a writable mapping exists is safe.
		if !errors.Is(restoreErr, windows.ERROR_SHARING_VIOLATION) && !errors.Is(restoreErr, windows.ERROR_LOCK_VIOLATION) && !errors.Is(restoreErr, windows.ERROR_ACCESS_DENIED) {
			t.Fatalf("unexpected mapped-source refusal: %v", restoreErr)
		}
		if r, found, err := c.GetQuarantine(id); err != nil || !found || r.Restored {
			t.Fatalf("refused restore state: %+v %v", r, err)
		}
		if err := windows.UnmapViewOfFile(view); err != nil {
			t.Fatal(err)
		}
		view = 0
		if err := windows.CloseHandle(mapping); err != nil {
			t.Fatal(err)
		}
		mapping = 0
		if err := s.Restore(id); err != nil {
			t.Fatalf("retry after mapping close: %v", err)
		}
		if b, err := os.ReadFile(original); err != nil || string(b) != "payload" {
			t.Fatalf("retry payload: %q %v", b, err)
		}
		return
	}
	// New admissions restore an independently verified private copy. A mapping
	// can prevent retirement of the old snapshot, but cannot mutate the output.
	writeOwnedRestoreMapping(t, view, []byte("changed"))
	if b, err := os.ReadFile(original); err != nil || string(b) != "payload" {
		t.Fatalf("writable mapping changed published snapshot: %q %v", b, err)
	}
	if _, err := os.Stat(stored); err != nil {
		t.Fatal("mapped source lost", err)
	}
	if r, found, err := c.GetQuarantine(id); err != nil || !found || !r.Restored {
		t.Fatalf("physical restore was not recorded: %+v %v", r, err)
	}
}

// Write only this test's mapped view in its own process; the OS owns the view,
// so it need not be converted from uintptr into an untracked Go pointer.
func writeOwnedRestoreMapping(t *testing.T, view uintptr, body []byte) {
	t.Helper()
	if view == 0 || len(body) != len("payload") {
		t.Fatal("invalid owned restore mapping span")
	}
	var written uintptr
	if err := windows.WriteProcessMemory(windows.CurrentProcess(), view, &body[0], uintptr(len(body)), &written); err != nil || written != uintptr(len(body)) {
		t.Fatalf("owned mapping write: bytes=%d error=%v", written, err)
	}
}
